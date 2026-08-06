package maintenance

import (
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/projsonal/gowms/internal/middleware"
	"github.com/projsonal/gowms/pkg/constant"
	"github.com/projsonal/gowms/pkg/utils"
)

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
	msg := "mode maintenance berhasil dinonaktifkan"
	if req.IsActive {
		msg = "mode maintenance berhasil diaktifkan"
	}
	return utils.OK(c, msg, toStatusResponse(status.IsActive, status.Message, status.StartedAt, status.EstimatedUntil))
}

func (h *Controller) RegisterRoutes(router fiber.Router) {
	g := router.Group("/maintenance")
	g.Get("/status", h.Status)
	g.Put("/", middleware.JWTAuth(h.jwtSvc), middleware.RequireRole(constant.RoleSuperAdmin), h.Set)
}
