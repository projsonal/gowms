package appinfo

import (
	"time"

	"github.com/gofiber/fiber/v2"

	notification "github.com/projsonal/gowms/internal/controller/notification"
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
	// SelfUpdateEnabled: false berarti fitur "Update Sekarang" belum
	// diaktifkan di server ini (lihat SelfUpdateConfig.Enabled) — frontend
	// pakai ini untuk memutuskan apakah tombol "Update Sekarang" boleh
	// ditampilkan sama sekali, atau cuma info "update tersedia, hubungi
	// administrator server" tanpa tombol aksi.
	SelfUpdateEnabled bool `json:"self_update_enabled"`
}

// CheckUpdate GET /app/check-update — bandingkan CurrentVersion (dibakar
// ke binary saat build) dengan rilis TERBARU di GitHub. Murni baca data,
// tidak mengubah apa pun di server — aman dipanggil kapan saja oleh
// siapa pun yang sudah login.
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

// UpdateStatus GET /app/update-status — status proses update TERAKHIR
// (idle kalau belum pernah ada). Dibaca dari FILE (bukan memori proses)
// karena proses backend sengaja restart di tengah alur update — lihat
// catatan panjang di pkg/selfupdate/status.go soal kenapa.
func (h *Controller) UpdateStatus(c *fiber.Ctx) error {
	status, err := selfupdate.ReadStatus(h.cfg.SelfUpdate.StatusPath)
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal membaca status update", nil)
	}

	// Begitu proses (baru, hasil restart, ATAU proses lama yang masih
	// hidup karena build gagal) sempat melihat status "success"/"failed"
	// yang belum ditindaklanjuti, matikan Mode Pemeliharaan OTOMATIS —
	// TAPI HANYA kalau kita sendiri yang menyalakannya (MaintenanceAuto),
	// supaya tidak diam-diam mematikan Mode Pemeliharaan yang memang
	// sengaja dinyalakan manual oleh admin lain untuk keperluan lain.
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

// TriggerUpdate POST /app/update — KHUSUS super_admin (lihat
// RegisterRoutes). Memvalidasi ulang ke GitHub (TIDAK percaya versi
// kiriman klien — mencegah klien memaksa "downgrade" ke tag sembarangan),
// menyalakan Mode Pemeliharaan, lalu menjalankan skrip deploy di LATAR
// BELAKANG (proses terpisah, lihat launchUpdateScript) dan langsung
// merespons — tidak menunggu skrip selesai, karena skrip itu pada
// akhirnya akan me-restart proses backend ini sendiri (lihat
// deploy/scripts/self-update.sh), yang otomatis memutus response HTTP
// yang sedang menunggu kalau kita menunggunya di sini.
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
		// Skrip gagal DIJALANKAN sama sekali (mis. file tidak ada / tidak
		// executable) — beda dari skrip jalan tapi build-nya gagal.
		// Rapikan status & Mode Pemeliharaan supaya tidak nyangkut di
		// "running" selamanya.
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
