#!/usr/bin/env bash
#
# deploy/scripts/self-update.sh — dijalankan OTOMATIS oleh backend
# (internal/controller/appinfo/launch_update.go) begitu super_admin
# menekan "Update Sekarang". TIDAK dimaksudkan dijalankan manual sehari-
# hari, tapi aman dicoba manual untuk debugging (lihat docs/self-update-
# setup.md bagian "Uji Manual").
#
# Alur:
#   1. Unduh binary rilis <tag> dari GitHub Releases (+ checksum-nya)
#   2. Backup binary yang sedang jalan
#   3. Pasang binary baru secara ATOMIK (mv, bukan copy isi ke tempat lama)
#   4. Restart service systemd
#   5. Health-check; kalau gagal -> ROLLBACK ke binary lama & restart lagi
#   6. Tulis hasil akhirnya ke SELFUPDATE_STATUS_PATH (file JSON yang sama
#      yang dibaca GET /app/update-status) — lihat pkg/selfupdate/status.go
#      untuk skema field yang dipertahankan/diubah.
#
# Kenapa skrip terpisah (bukan logic Go semua)? Karena begitu binary baru
# terpasang, proses backend LAMA (yang memulai skrip ini) akan mati waktu
# `systemctl restart` dipanggil — supaya proses restart & pelaporan hasil
# akhir tetap jalan sampai selesai, itu HARUS berada di proses yang
# independen dari backend, yaitu skrip shell ini (dijalankan via Setsid,
# lihat launch_update.go).
#
# Env yang WAJIB diisi oleh pemanggil (launch_update.go):
#   SELFUPDATE_SERVICE_NAME   nama unit systemd, mis. gowms-backend
#   SELFUPDATE_WORKDIR        direktori kerja (juga cwd skrip ini)
#   SELFUPDATE_STATUS_PATH    path ABSOLUT file status JSON
#   SELFUPDATE_GITHUB_OWNER   owner repo GitHub, mis. projsonal
#   SELFUPDATE_GITHUB_REPO    nama repo GitHub, mis. gowms
# Argumen:
#   $1  tag versi tujuan, mis. v1.4.0 (sama dengan SELFUPDATE_TO_VERSION)
#
# Env OPSIONAL (boleh tidak diisi, ada default masuk akal):
#   SELFUPDATE_BIN_NAME   nama file binary di WorkDir (default: gowms-backend)
#   SELFUPDATE_HEALTH_URL URL health check (default: http://127.0.0.1:8080/health/live)
#   SELFUPDATE_GH_TOKEN   token GitHub (kalau repo privat / menghindari rate limit)
#   SELFUPDATE_RETRIES    jumlah percobaan poll health check (default: 10)
#   SELFUPDATE_RETRY_DELAY jeda antar percobaan dalam detik (default: 3)

set -Eeuo pipefail

# ---------------------------------------------------------------------
# 0. Validasi input & siapkan variabel
# ---------------------------------------------------------------------
TO_VERSION="${1:?Usage: self-update.sh <tag>}"
SERVICE_NAME="${SELFUPDATE_SERVICE_NAME:?SELFUPDATE_SERVICE_NAME wajib diisi}"
WORKDIR="${SELFUPDATE_WORKDIR:?SELFUPDATE_WORKDIR wajib diisi}"
STATUS_PATH="${SELFUPDATE_STATUS_PATH:?SELFUPDATE_STATUS_PATH wajib diisi}"
GITHUB_OWNER="${SELFUPDATE_GITHUB_OWNER:?SELFUPDATE_GITHUB_OWNER wajib diisi}"
GITHUB_REPO="${SELFUPDATE_GITHUB_REPO:?SELFUPDATE_GITHUB_REPO wajib diisi}"

BIN_NAME="${SELFUPDATE_BIN_NAME:-gowms-backend}"
HEALTH_URL="${SELFUPDATE_HEALTH_URL:-http://127.0.0.1:8080/health/live}"
GH_TOKEN="${SELFUPDATE_GH_TOKEN:-}"
RETRIES="${SELFUPDATE_RETRIES:-10}"
RETRY_DELAY="${SELFUPDATE_RETRY_DELAY:-3}"

cd "$WORKDIR"

BIN_PATH="${WORKDIR}/${BIN_NAME}"
NEW_BIN_PATH="${BIN_PATH}.new"
BAK_BIN_PATH="${BIN_PATH}.bak"
ASSET_NAME="${BIN_NAME}-linux-amd64"
CHECKSUM_NAME="${ASSET_NAME}.sha256"
RELEASE_BASE_URL="https://github.com/${GITHUB_OWNER}/${GITHUB_REPO}/releases/download/${TO_VERSION}"

# Ditandai true begitu binary lama sudah dipindah ke .bak DAN binary baru
# sudah dipasang menggantikan BIN_PATH — dipakai on_error untuk memutuskan
# apakah perlu ROLLBACK (kalau gagal SEBELUM titik ini, tidak ada apa pun
# yang perlu dikembalikan, binary lama masih utuh di tempatnya).
SWAPPED=false

log() {
	printf '[%s] %s\n' "$(date -u +%Y-%m-%dT%H:%M:%SZ)" "$*"
}

require_cmd() {
	if ! command -v "$1" >/dev/null 2>&1; then
		log "FATAL: perintah '$1' tidak ditemukan di PATH — lihat docs/self-update-setup.md (Prasyarat)"
		exit 1
	fi
}

# ---------------------------------------------------------------------
# write_status STATE MESSAGE — tulis/patch file status JSON secara
# ATOMIK (tulis ke .tmp lalu mv), sama seperti pkg/selfupdate/status.go
# WriteStatus. Kalau file sudah ada (SELALU ada di titik ini, ditulis
# TriggerUpdate sebelum skrip ini dijalankan), field from_version/
# to_version/started_at/maintenance_auto DIPERTAHANKAN apa adanya —
# skrip ini HANYA berhak mengubah state/message/finished_at, supaya
# proses backend yang membaca status setelah restart tetap tahu apakah
# boleh mematikan Mode Pemeliharaan otomatis (field maintenance_auto).
# ---------------------------------------------------------------------
write_status() {
	local state="$1" message="$2"
	local finished_at tmp
	finished_at="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
	tmp="${STATUS_PATH}.tmp"
	mkdir -p "$(dirname "$STATUS_PATH")"

	if [ -f "$STATUS_PATH" ]; then
		jq --arg state "$state" --arg message "$message" --arg finished_at "$finished_at" \
			'.state = $state | .message = $message | .finished_at = $finished_at' \
			"$STATUS_PATH" >"$tmp"
	else
		# Semestinya tidak pernah terjadi (TriggerUpdate selalu menulis
		# status "running" dulu) — fallback ini murni jaga-jaga supaya
		# skrip tetap melaporkan hasil walau file awal entah kenapa hilang.
		jq -n --arg state "$state" --arg message "$message" \
			--arg finished_at "$finished_at" --arg to "$TO_VERSION" \
			'{state: $state, message: $message, to_version: $to, finished_at: $finished_at, maintenance_auto: false, acknowledged: false}' \
			>"$tmp"
	fi
	mv "$tmp" "$STATUS_PATH"
	log "status ditulis: state=${state} message=${message}"
}

restart_service() {
	log "me-restart service ${SERVICE_NAME}..."
	sudo systemctl restart "$SERVICE_NAME"
}

# health_check — poll HEALTH_URL sampai RETRIES kali, jeda RETRY_DELAY
# detik. Return 0 kalau service sehat, 1 kalau tidak sampai batas
# percobaan habis.
health_check() {
	local i
	for ((i = 1; i <= RETRIES; i++)); do
		if systemctl is-active --quiet "$SERVICE_NAME" &&
			curl -fsS --max-time 5 "$HEALTH_URL" >/dev/null 2>&1; then
			log "health check OK (percobaan ${i}/${RETRIES})"
			return 0
		fi
		log "health check belum sehat (percobaan ${i}/${RETRIES}), tunggu ${RETRY_DELAY}s..."
		sleep "$RETRY_DELAY"
	done
	return 1
}

# rollback — kembalikan binary lama & restart ulang. Dipanggil dari
# on_error KALAU swap sempat terjadi. Kegagalan di sini sendiri tidak
# ditangani lebih lanjut (skrip sudah dalam kondisi gagal) — cukup dicatat
# ke log supaya admin tahu perlu turun tangan manual.
rollback() {
	if [ ! -f "$BAK_BIN_PATH" ]; then
		log "PERINGATAN: tidak bisa rollback, backup ${BAK_BIN_PATH} tidak ada"
		return
	fi
	log "melakukan ROLLBACK ke binary sebelumnya..."
	mv -f "$BAK_BIN_PATH" "$BIN_PATH"
	if restart_service; then
		if health_check; then
			log "rollback berhasil, service kembali sehat dengan binary lama"
		else
			log "PERINGATAN: rollback selesai tapi health check tetap gagal — perlu pemeriksaan manual"
		fi
	else
		log "PERINGATAN: restart service gagal saat proses rollback — perlu pemeriksaan manual"
	fi
}

on_error() {
	local exit_code=$?
	log "GAGAL pada baris ${BASH_LINENO[0]} (perintah: ${BASH_COMMAND}) — exit code ${exit_code}"
	if [ "$SWAPPED" = true ]; then
		rollback
		write_status "failed" "Update ke ${TO_VERSION} gagal (${BASH_COMMAND}) — sistem dikembalikan otomatis ke versi sebelumnya."
	else
		write_status "failed" "Update ke ${TO_VERSION} gagal sebelum binary diganti (${BASH_COMMAND}) — versi sebelumnya tetap berjalan tanpa perubahan."
	fi
	exit "$exit_code"
}
trap on_error ERR

# ---------------------------------------------------------------------
# 1. Prasyarat
# ---------------------------------------------------------------------
require_cmd curl
require_cmd jq
require_cmd sha256sum
require_cmd systemctl
require_cmd sudo

log "mulai update ke ${TO_VERSION} (service=${SERVICE_NAME}, workdir=${WORKDIR})"

CURL_AUTH_ARGS=()
if [ -n "$GH_TOKEN" ]; then
	CURL_AUTH_ARGS=(-H "Authorization: Bearer ${GH_TOKEN}")
fi

# ---------------------------------------------------------------------
# 2. Unduh binary rilis baru + checksum-nya
# ---------------------------------------------------------------------
log "mengunduh ${ASSET_NAME} dari rilis ${TO_VERSION}..."
curl -fsSL --retry 3 --retry-delay 2 --max-time 120 \
	"${CURL_AUTH_ARGS[@]}" \
	-o "$NEW_BIN_PATH" \
	"${RELEASE_BASE_URL}/${ASSET_NAME}"

if [ ! -s "$NEW_BIN_PATH" ]; then
	log "FATAL: file yang terunduh kosong"
	exit 1
fi

log "mengunduh checksum ${CHECKSUM_NAME}..."
curl -fsSL --retry 3 --retry-delay 2 --max-time 30 \
	"${CURL_AUTH_ARGS[@]}" \
	-o "${NEW_BIN_PATH}.sha256" \
	"${RELEASE_BASE_URL}/${CHECKSUM_NAME}"

log "memverifikasi checksum..."
EXPECTED_HASH="$(awk '{print $1}' "${NEW_BIN_PATH}.sha256")"
ACTUAL_HASH="$(sha256sum "$NEW_BIN_PATH" | awk '{print $1}')"
if [ "$EXPECTED_HASH" != "$ACTUAL_HASH" ]; then
	log "FATAL: checksum tidak cocok (expected=${EXPECTED_HASH} actual=${ACTUAL_HASH}) — binary yang terunduh mungkin korup/di-tampering"
	exit 1
fi
rm -f "${NEW_BIN_PATH}.sha256"
chmod +x "$NEW_BIN_PATH"

# ---------------------------------------------------------------------
# 3. Backup binary lama & pasang binary baru secara ATOMIK
# ---------------------------------------------------------------------
log "backup binary saat ini ke ${BAK_BIN_PATH}..."
cp -f "$BIN_PATH" "$BAK_BIN_PATH"

log "memasang binary baru (mv atomik)..."
mv -f "$NEW_BIN_PATH" "$BIN_PATH"
chmod +x "$BIN_PATH"
SWAPPED=true

# ---------------------------------------------------------------------
# 4. Restart & health check
# ---------------------------------------------------------------------
restart_service

log "menunggu service pulih sebelum health check pertama..."
sleep 2

if ! health_check; then
	log "health check gagal setelah ${RETRIES} percobaan"
	# Biarkan `set -e`/trap ERR yang menangani rollback+status supaya
	# hanya ada SATU jalur pelaporan kegagalan (on_error) — konsisten
	# dengan semua kegagalan lain di skrip ini.
	false
fi

# ---------------------------------------------------------------------
# 5. Sukses — bersihkan & laporkan
# ---------------------------------------------------------------------
rm -f "$BAK_BIN_PATH"
write_status "success" "Update ke ${TO_VERSION} berhasil, service berjalan normal."
log "update ke ${TO_VERSION} SELESAI dengan sukses"
