//go:build !windows

package appinfo

import (
	"os/exec"
	"syscall"
)

// detachProcess (Unix/Linux) — jadikan skrip deploy pemimpin sesi/process
// group baru (Setsid) supaya TIDAK ikut mati kalau proses backend yang
// memulainya di-restart systemd di tengah eksekusi skrip. Ini yang benar-
// benar dipakai di produksi (VPS selalu Linux — lihat deploy/scripts/
// self-update.sh & workflows/backend-deploy.yml).
func detachProcess(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
}
