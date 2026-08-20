# Panduan Setup: Cek Update / Update Sekarang (Self-Update via GitHub Releases)

Fitur ini ada di **Settings > Sistem** dan terdiri dari 3 endpoint (semua wajib
login, `POST /app/update` khusus `super_admin`):

| Endpoint | Fungsi |
|---|---|
| `GET /app/check-update` | Bandingkan versi berjalan vs rilis GitHub terbaru |
| `GET /app/update-status` | Status proses update terakhir (idle/running/success/failed) |
| `POST /app/update` | Mulai proses update (khusus `super_admin`) |

Alur singkatnya: backend mengecek rilis terbaru di GitHub → kalau ada versi
lebih baru & `super_admin` menekan "Update Sekarang" → backend menyalakan
Mode Pemeliharaan → menjalankan `deploy/scripts/self-update.sh` di latar
belakang → skrip mengunduh binary rilis, memasangnya, me-restart service
systemd, health-check → hasilnya ditulis ke file status yang dibaca
`/app/update-status` (proses backend baru hasil restart tetap bisa membacanya
karena statusnya di file, bukan di memori — lihat komentar panjang di
`pkg/selfupdate/status.go`).

File yang relevan:
- `internal/controller/appinfo/update_controller.go` — handler HTTP
- `internal/controller/appinfo/launch_update.go` — menjalankan skrip deploy
- `pkg/selfupdate/github.go` — client GitHub Releases API + pembanding versi
- `pkg/selfupdate/status.go` — skema & baca/tulis file status
- `deploy/scripts/self-update.sh` — skrip yang benar-benar swap binary + restart
- `workflows/backend-release.yml` — CI yang mem-publish rilis GitHub

---

## 1. Prasyarat di VPS

```bash
# jq wajib — dipakai self-update.sh untuk patch file status JSON
sudo apt-get update && sudo apt-get install -y jq curl coreutils
```

Pastikan juga `sudo`, `systemctl`, `sha256sum` tersedia (biasanya sudah ada
di distro server pada umumnya).

## 2. Izinkan service user me-restart systemd TANPA password

Backend berjalan sebagai user tertentu (mis. `gowms`, sesuai unit systemd).
User itu perlu bisa menjalankan `systemctl restart <service>` DAN
`systemctl is-active <service>` tanpa diminta password interaktif, karena
`self-update.sh` dijalankan tanpa terminal.

```bash
sudo visudo -f /etc/sudoers.d/gowms-selfupdate
```

Isi (sesuaikan `gowms` dengan user service sebenarnya, dan `gowms-backend`
dengan `AUTO_UPDATE_SERVICE_NAME`):

```
gowms ALL=(root) NOPASSWD: /usr/bin/systemctl restart gowms-backend
gowms ALL=(root) NOPASSWD: /usr/bin/systemctl is-active gowms-backend
```

Jangan pakai wildcard (`systemctl *`) — batasi persis ke unit ini supaya
kompromi pada backend tidak otomatis berarti kompromi seluruh server.

## 3. Siapkan skrip deploy

```bash
cd /opt/gowms/backend
mkdir -p deploy/scripts
# salin deploy/scripts/self-update.sh dari repo ke sini
chmod +x deploy/scripts/self-update.sh
```

`launchUpdateScript` (lihat `launch_update.go`) menolak menjalankan skrip
yang tidak executable — pesan errornya akan menyebutkan dokumen ini kalau
langkah ini terlewat.

## 4. Environment variable backend

Tambahkan ke `.env` (atau environment unit systemd) di server produksi:

```bash
AUTO_UPDATE_ENABLED=true
AUTO_UPDATE_GITHUB_OWNER=projsonal
AUTO_UPDATE_GITHUB_REPO=gowms
AUTO_UPDATE_SCRIPT_PATH=./deploy/scripts/self-update.sh
AUTO_UPDATE_WORKDIR=/opt/gowms/backend
AUTO_UPDATE_SERVICE_NAME=gowms-backend
AUTO_UPDATE_STATUS_PATH=./var/run/self-update-status.json
```

Catatan:
- `AUTO_UPDATE_WORKDIR` **harus** sama dengan `WorkingDirectory` di unit
  systemd backend — supaya path relatif (mis. `AUTO_UPDATE_STATUS_PATH`)
  konsisten dibaca oleh proses lama maupun proses baru hasil restart.
- Selama `AUTO_UPDATE_ENABLED` belum `true`, `GET /app/check-update` dan
  `GET /app/update-status` tetap jalan normal (murni baca) — cuma
  `POST /app/update` yang ditolak dengan pesan mengarahkan ke dokumen ini.
  Ini sengaja: admin bisa lihat dulu "ada update tersedia" sebelum server
  benar-benar disiapkan untuk auto-update.
- Repo GitHub **harus publik**, atau isi `SELFUPDATE_GH_TOKEN` (lihat §6)
  kalau privat.

## 5. Siapkan alur rilis (tag → GitHub Release)

Fitur ini bergantung pada **GitHub Release** yang benar-benar ada (bukan
sekadar git tag) — endpoint `releases/latest` GitHub hanya mengembalikan
sesuatu kalau rilis resmi pernah dipublikasikan, dan `self-update.sh`
mengunduh binary dari asset rilis tersebut.

1. Salin `workflows/backend-release.yml` ke `.github/workflows/release.yml`
   di repo (workflow ini terpisah dari `deploy.yml` yang sudah ada — lihat
   komentar di kedua file untuk perbedaannya).
2. Setiap kali siap rilis versi baru:
   ```bash
   # update CurrentVersion di internal/controller/appinfo/appinfo_controller.go
   # + tambah entri baru di changelogData & changelog.yml
   git commit -am "chore: rilis v1.4.0"
   git push origin main
   git tag v1.4.0
   git push origin v1.4.0
   ```
3. Workflow otomatis build `gowms-backend-linux-amd64` +
   `gowms-backend-linux-amd64.sha256`, lalu publish sebagai GitHub Release
   dengan tag yang sama.

**Penting:** nilai tag (`v1.4.0`) harus bisa dibandingkan apa adanya dengan
`CurrentVersion` yang dibakar ke binary lewat `selfupdate.CompareVersions`
(lihat `pkg/selfupdate/github.go` untuk aturan toleransinya).

## 6. (Opsional) Repo privat

Kalau `projsonal/gowms` privat, `FetchLatestRelease` (Go) tetap butuh token
kalau nanti mau mendukung repo privat (belum diimplementasikan — lihat
komentar di `pkg/selfupdate/github.go`), dan `self-update.sh` butuh:

```bash
AUTO_UPDATE... # tidak ada var baru di config.go untuk ini
```

Untuk sekarang, cara termudah kalau repo privat: set token langsung di
environment unit systemd (bukan lewat `pkg/config`, supaya token tidak
tercampur dengan config aplikasi umum):

```
Environment=SELFUPDATE_GH_TOKEN=ghp_xxx
```

`self-update.sh` sudah membaca `SELFUPDATE_GH_TOKEN` kalau ada dan
mengirim header `Authorization: Bearer` saat mengunduh asset.

## 7. Uji manual

Sebelum mengandalkan tombol "Update Sekarang" di produksi, jalankan skrip
manual dulu di server staging dengan versi tag yang SUDAH ada rilisnya:

```bash
cd /opt/gowms/backend
SELFUPDATE_SERVICE_NAME=gowms-backend \
SELFUPDATE_WORKDIR=/opt/gowms/backend \
SELFUPDATE_STATUS_PATH=/opt/gowms/backend/var/run/self-update-status.json \
SELFUPDATE_GITHUB_OWNER=projsonal \
SELFUPDATE_GITHUB_REPO=gowms \
  ./deploy/scripts/self-update.sh v1.4.0

cat var/run/self-update-status.json
```

Periksa juga log detailnya di `var/log/self-update-v1.4.0.log`.

## 8. Perilaku kegagalan & rollback

- **Gagal sebelum binary diganti** (mis. gagal unduh, checksum tidak
  cocok): binary lama tidak disentuh, service tidak di-restart, status
  ditulis `failed`.
- **Gagal setelah binary diganti** (restart gagal / health check gagal
  setelah `SELFUPDATE_RETRIES` percobaan): skrip otomatis mengembalikan
  binary lama dari backup (`gowms-backend.bak`) dan me-restart service
  sekali lagi, baru menulis status `failed` dengan catatan bahwa rollback
  sudah dilakukan.
- **Sukses**: backup dihapus, status ditulis `success`. Proses backend LAMA
  (yang menerima request `POST /app/update`) sudah lama mati kena restart —
  proses BARU yang membaca `GET /app/update-status` pertama kali akan
  otomatis mematikan Mode Pemeliharaan (kalau dinyalakan otomatis oleh
  proses update ini) dan mengirim notifikasi ke semua `super_admin` (lihat
  `UpdateStatus` di `update_controller.go`).

## 9. Checklist ringkas

- [ ] `jq`, `curl`, `sha256sum`, `sudo`, `systemctl` tersedia di VPS
- [ ] sudoers NOPASSWD untuk `systemctl restart/is-active <service>` ini saja
- [ ] `deploy/scripts/self-update.sh` ada di `AUTO_UPDATE_WORKDIR` & executable
- [ ] Env `AUTO_UPDATE_*` terisi, `AUTO_UPDATE_ENABLED=true`
- [ ] `.github/workflows/release.yml` terpasang di repo
- [ ] Sudah pernah tag + push minimal satu rilis, dan `gowms-backend-linux-amd64`
      + `.sha256`-nya muncul di halaman Release GitHub
- [ ] Sudah dicoba manual (§7) di staging sebelum dipakai di produksi
