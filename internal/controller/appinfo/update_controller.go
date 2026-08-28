package appinfo

import (
	"time"

	"github.com/gofiber/fiber/v2"

	notification "github.com/projsonal/gowms/internal/controller/notifikasi"
	"github.com/projsonal/gowms/pkg/constant"
	"github.com/projsonal/gowms/pkg/selfupdate"
	"github.com/projsonal/gowms/pkg/utils"
)

type CheckUpdateResponse struct {
	CurrentVersion  string `json:"current_version"`
	LatestVersion   string `json:"latest_version,omitempty"`
	UpdateAvailable bool   `json:"update_available"`
	ReleaseURL      string `json:"release_url,omitempty"`
	ReleaseNotes    string `json:"release_notes,omitempty"`
	PublishedAt     string `json:"published_at,omitempty"`

	SelfUpdateEnabled bool `json:"self_update_enabled"`
}

func (h *Controller) CheckUpdate(c *fiber.Ctx) error {
	release, err := selfupdate.FetchLatestRelease(h.cfg.SelfUpdate.GitHubOwner, h.cfg.SelfUpdate.GitHubRepo)
	if err != nil {
		if err == selfupdate.ErrNoReleases {
			return utils.OK(c, "belum ada rilis resmi di GitHub", CheckUpdateResponse{
				CurrentVersion:    CurrentVersion,
				UpdateAvailable:   false,
				SelfUpdateEnabled: h.cfg.SelfUpdate.Enabled,
			})
		}
		return utils.Fail(c, fiber.StatusBadGateway, "gagal mengecek update: "+err.Error(), nil)
	}

	updateAvailable := selfupdate.CompareVersions(release.TagName, CurrentVersion) > 0
	return utils.OK(c, "berhasil mengecek update", CheckUpdateResponse{
		CurrentVersion:    CurrentVersion,
		LatestVersion:     release.TagName,
		UpdateAvailable:   updateAvailable,
		ReleaseURL:        release.HTMLURL,
		ReleaseNotes:      release.Body,
		PublishedAt:       release.PublishedAt,
		SelfUpdateEnabled: h.cfg.SelfUpdate.Enabled,
	})
}

func (h *Controller) UpdateStatus(c *fiber.Ctx) error {
	status, err := selfupdate.ReadStatus(h.cfg.SelfUpdate.StatusPath)
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal membaca status update", nil)
	}

	if !status.Acknowledged && (status.State == selfupdate.StateSuccess || status.State == selfupdate.StateFailed) {
		if status.MaintenanceAuto {
			_, _ = h.maintenanceRepo.Set(false, "", nil, 0)
		}
		status.Acknowledged = true
		_ = selfupdate.WriteStatus(h.cfg.SelfUpdate.StatusPath, status)

		notifTitle := "Update Aplikasi Berhasil"
		notifMsg := "Aplikasi berhasil diperbarui ke " + status.ToVersion + "."
		if status.State == selfupdate.StateFailed {
			notifTitle = "Update Aplikasi Gagal"
			notifMsg = "Percobaan update ke " + status.ToVersion + " gagal: " + status.Message + ". Versi sebelumnya tetap berjalan."
		}
		notifyRoleAdmin(h, notifTitle, notifMsg)
	}

	return utils.OK(c, "status update berhasil diambil", status)
}

func (h *Controller) TriggerUpdate(c *fiber.Ctx) error {
	if !h.cfg.SelfUpdate.Enabled {
		return utils.Fail(c, fiber.StatusForbidden,
			"fitur Update Sekarang belum diaktifkan di server ini — set AUTO_UPDATE_ENABLED=true setelah mengikuti docs/self-update-setup.md", nil)
	}

	current, err := selfupdate.ReadStatus(h.cfg.SelfUpdate.StatusPath)
	if err == nil && current.State == selfupdate.StateRunning {
		return utils.Fail(c, fiber.StatusConflict, "proses update lain sedang berjalan, tunggu sampai selesai", nil)
	}

	release, err := selfupdate.FetchLatestRelease(h.cfg.SelfUpdate.GitHubOwner, h.cfg.SelfUpdate.GitHubRepo)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadGateway, "gagal mengecek rilis terbaru: "+err.Error(), nil)
	}
	if selfupdate.CompareVersions(release.TagName, CurrentVersion) <= 0 {
		return utils.Fail(c, fiber.StatusBadRequest, "sudah versi terbaru, tidak ada yang perlu diupdate", nil)
	}

	maintenanceStatus, _ := h.maintenanceRepo.Get()
	wasAlreadyActive := maintenanceStatus != nil && maintenanceStatus.IsActive
	if !wasAlreadyActive {
		userID, _ := c.Locals(constant.CtxUserID).(uint)
		if _, err := h.maintenanceRepo.Set(true, "Sedang memperbarui aplikasi ke versi "+release.TagName+"...", nil, userID); err != nil {
			return utils.Fail(c, fiber.StatusInternalServerError, "gagal mengaktifkan Mode Pemeliharaan", nil)
		}
	}

	now := time.Now()
	status := selfupdate.Status{
		State:           selfupdate.StateRunning,
		Message:         "Menjalankan skrip deploy...",
		FromVersion:     CurrentVersion,
		ToVersion:       release.TagName,
		StartedAt:       &now,
		MaintenanceAuto: !wasAlreadyActive,
	}
	if err := selfupdate.WriteStatus(h.cfg.SelfUpdate.StatusPath, status); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal menyimpan status update", nil)
	}

	if err := launchUpdateScript(updateScriptParams{
		ScriptPath:  h.cfg.SelfUpdate.ScriptPath,
		WorkDir:     h.cfg.SelfUpdate.WorkDir,
		ServiceName: h.cfg.SelfUpdate.ServiceName,
		ToVersion:   release.TagName,
		StatusPath:  h.cfg.SelfUpdate.StatusPath,
		GitHubOwner: h.cfg.SelfUpdate.GitHubOwner,
		GitHubRepo:  h.cfg.SelfUpdate.GitHubRepo,
	}); err != nil {

		status.State = selfupdate.StateFailed
		status.Message = "Gagal menjalankan skrip deploy: " + err.Error()
		finished := time.Now()
		status.FinishedAt = &finished
		_ = selfupdate.WriteStatus(h.cfg.SelfUpdate.StatusPath, status)
		if status.MaintenanceAuto {
			_, _ = h.maintenanceRepo.Set(false, "", nil, 0)
		}
		return utils.Fail(c, fiber.StatusInternalServerError, status.Message, nil)
	}

	return utils.OK(c, "update dimulai di latar belakang — pantau progresnya lewat /app/update-status", status)
}

func notifyRoleAdmin(h *Controller, title, message string) {
	notification.Notify(h.notifRepo, "app_update", title, message, "/settings", nil, constant.RoleSuperAdmin)
}
