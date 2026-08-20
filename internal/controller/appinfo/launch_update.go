package appinfo

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

// updateScriptParams — semua yang dibutuhkan skrip deploy
// (deploy/scripts/self-update.sh) untuk mengunduh rilis GitHub yang benar
// & menuliskan status akhirnya ke file yang sama yang dibaca UpdateStatus.
type updateScriptParams struct {
	ScriptPath  string
	WorkDir     string
	ServiceName string
	ToVersion   string
	// StatusPath diteruskan sebagai path ABSOLUT (lihat launchUpdateScript)
	// supaya skrip tetap menulis ke file yang sama persis dengan yang
	// dibaca proses backend, terlepas dari cwd skrip (cmd.Dir = WorkDir,
	// yang belum tentu sama dengan cwd proses backend saat StatusPath
	// relatif di-resolve pertama kali oleh config.Load()).
	StatusPath  string
	GitHubOwner string
	GitHubRepo  string
}

// launchUpdateScript menjalankan skrip deploy DI PROSES TERPISAH, tidak
// ditunggu (fire-and-forget) — lihat catatan panjang di TriggerUpdate
// kenapa: skrip pada akhirnya me-restart proses backend ini sendiri lewat
// systemd, yang akan memutus apa pun yang masih menunggu proses ini kalau
// kita Wait() di sini.
//
// Setsid dipakai supaya skrip TIDAK ikut mati waktu proses induk (backend
// ini) di-restart systemd di tengah eksekusi skrip — kalau skrip masih
// jadi anak proses langsung backend lama, sinyal yang dikirim systemd ke
// seluruh process group backend lama bisa ikut membunuh skrip yang
// sedang berjalan sebelum sempat menyelesaikan swap binary.
//
// Output skrip (stdout+stderr) dialihkan ke file log terpisah
// (var/log/self-update-<tag>.log) supaya bisa diperiksa manual kalau
// UpdateStatus tidak cukup detail (mis. skrip gagal SEBELUM sempat
// menulis file status sama sekali).
func launchUpdateScript(p updateScriptParams) error {
	absScript, err := filepath.Abs(p.ScriptPath)
	if err != nil {
		return fmt.Errorf("path skrip tidak valid: %w", err)
	}
	info, err := os.Stat(absScript)
	if err != nil {
		return fmt.Errorf("skrip deploy tidak ditemukan di %s: %w", absScript, err)
	}
	if info.IsDir() || info.Mode()&0o111 == 0 {
		return fmt.Errorf("skrip deploy di %s tidak executable — jalankan chmod +x (lihat docs/self-update-setup.md)", absScript)
	}

	absStatusPath, err := filepath.Abs(p.StatusPath)
	if err != nil {
		return fmt.Errorf("path status tidak valid: %w", err)
	}

	logDir := filepath.Join("var", "log")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return fmt.Errorf("gagal membuat folder log: %w", err)
	}
	logFile, err := os.OpenFile(
		filepath.Join(logDir, "self-update-"+p.ToVersion+".log"),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644,
	)
	if err != nil {
		return fmt.Errorf("gagal membuka file log update: %w", err)
	}

	cmd := exec.Command(absScript, p.ToVersion)
	cmd.Dir = p.WorkDir
	cmd.Env = append(os.Environ(),
		"SELFUPDATE_SERVICE_NAME="+p.ServiceName,
		"SELFUPDATE_WORKDIR="+p.WorkDir,
		"SELFUPDATE_TO_VERSION="+p.ToVersion,
		"SELFUPDATE_STATUS_PATH="+absStatusPath,
		"SELFUPDATE_GITHUB_OWNER="+p.GitHubOwner,
		"SELFUPDATE_GITHUB_REPO="+p.GitHubRepo,
	)
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	// detachProcess: jadikan skrip pemimpin sesi/process group baru,
	// terlepas dari process group proses backend ini — lihat catatan di
	// atas. Implementasinya beda per-OS (field syscall.SysProcAttr TIDAK
	// sama antara Unix & Windows — mis. Setsid cuma ada di Unix), jadi
	// dipisah ke launch_update_unix.go / launch_update_windows.go supaya
	// paket ini tetap bisa di-build di mesin dev Windows sekalipun target
	// deploy sebenarnya SELALU Linux (systemd, lihat docs/self-update-
	// setup.md).
	detachProcess(cmd)

	if err := cmd.Start(); err != nil {
		_ = logFile.Close()
		return fmt.Errorf("gagal memulai proses skrip: %w", err)
	}

	// Lepas proses skrip sepenuhnya — TIDAK Wait() secara sinkron di alur
	// request, biar skrip lanjut jalan independen. Goroutine ini cuma
	// menunggu supaya file descriptor logFile ditutup rapi begitu skrip
	// (atau proses backend baru hasil restart) selesai — TIDAK
	// memengaruhi response HTTP yang sudah lebih dulu dikirim oleh
	// TriggerUpdate.
	go func() {
		_ = cmd.Wait()
		_ = logFile.Close()
	}()

	log.Printf("selfupdate: skrip deploy dimulai (pid=%d) -> %s (target versi %s)", cmd.Process.Pid, absScript, p.ToVersion)
	return nil
}
