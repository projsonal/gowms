package maintenance

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/gofiber/fiber/v2"

	notification "github.com/projsonal/gowms/internal/controller/notification"
	"github.com/projsonal/gowms/internal/middleware"
	"github.com/projsonal/gowms/pkg/constant"
	"github.com/projsonal/gowms/pkg/utils"
)

// logMaintenanceEvent menulis satu baris audit trail ke var/log/maintenance.log
// setiap kali super_admin mengaktifkan/menonaktifkan mode maintenance —
// terpisah dari backend.log umum supaya gampang diaudit siapa & kapan
// fitur ini dipakai (dampaknya besar: memblokir semua role non-super_admin).
func logMaintenanceEvent(userID uint, isActive bool, message string) {
	logDir := filepath.Join("var", "log")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		log.Printf("maintenance: gagal membuat folder log: %v", err)
		return
	}
	f, err := os.OpenFile(filepath.Join(logDir, "maintenance.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		log.Printf("maintenance: gagal membuka maintenance.log: %v", err)
		return
	}
	defer func() { _ = f.Close() }()

	action := "DINONAKTIFKAN"
	if isActive {
		action = "DIAKTIFKAN"
	}
	line := fmt.Sprintf("[%s] mode maintenance %s oleh user_id=%d — pesan: %q\n",
		time.Now().Format(time.RFC3339), action, userID, message)
	if _, err := f.WriteString(line); err != nil {
		log.Printf("maintenance: gagal menulis ke maintenance.log: %v", err)
	}
}

func toStatusResponse(active bool, message string, startedAt, estimatedUntil *time.Time) StatusResponse {
	res := StatusResponse{IsActive: active, Message: message, StartedAt: startedAt, EstimatedUntil: estimatedUntil}
	if active && estimatedUntil != nil {
		if remaining := time.Until(*estimatedUntil); remaining > 0 {
			res.RemainingSeconds = int64(remaining.Seconds())
		}
	}
	return res
}

func (h *Controller) Status(c *fiber.Ctx) error {
	status, err := h.repo.Get()
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal mengambil status maintenance", nil)
	}
	return utils.OK(c, "status maintenance berhasil diambil",
		toStatusResponse(status.IsActive, status.Message, status.StartedAt, status.EstimatedUntil))
}

func (h *Controller) Set(c *fiber.Ctx) error {
	var req SetRequest
	if !utils.ParseAndValidate(c, &req) {
		return nil
	}
	userID, _ := c.Locals(constant.CtxUserID).(uint)

	status, err := h.repo.Set(req.IsActive, req.Message, req.EstimatedUntil, userID)
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal memperbarui status maintenance", nil)
	}
	logMaintenanceEvent(userID, req.IsActive, req.Message)
	msg := "mode maintenance berhasil dinonaktifkan"
	title := "Mode Pemeliharaan Dinonaktifkan"
	body := "Sistem kembali normal, semua fitur bisa diakses seperti biasa."
	if req.IsActive {
		msg = "mode maintenance berhasil diaktifkan"
		title = "Mode Pemeliharaan Diaktifkan"
		body = "Sebagian fitur untuk sementara tidak bisa diakses."
		if req.Message != "" {
			body = req.Message
		}
	}
	notification.Notify(h.notifRepo, "maintenance", title, body, "", nil, "all")
	return utils.OK(c, msg, toStatusResponse(status.IsActive, status.Message, status.StartedAt, status.EstimatedUntil))
}

func (h *Controller) RegisterRoutes(router fiber.Router) {
	g := router.Group("/maintenance")
	g.Get("/status", h.Status)
	g.Put("/", middleware.JWTAuth(h.jwtSvc), middleware.RequireRole(constant.RoleSuperAdmin), h.Set)
}
