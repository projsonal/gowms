package appinfo

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

type updateScriptParams struct {
	ScriptPath  string
	WorkDir     string
	ServiceName string
	ToVersion   string

	StatusPath  string
	GitHubOwner string
	GitHubRepo  string
}

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

	detachProcess(cmd)

	if err := cmd.Start(); err != nil {
		_ = logFile.Close()
		return fmt.Errorf("gagal memulai proses skrip: %w", err)
	}

	go func() {
		_ = cmd.Wait()
		_ = logFile.Close()
	}()

	log.Printf("selfupdate: skrip deploy dimulai (pid=%d) -> %s (target versi %s)", cmd.Process.Pid, absScript, p.ToVersion)
	return nil
}
