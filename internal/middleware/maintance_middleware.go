package middleware

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/projsonal/gowms/internal/model"
	"github.com/projsonal/gowms/pkg/constant"
	"github.com/projsonal/gowms/pkg/utils"
)

type MaintenancePayload struct {
	Maintenance      bool   `json:"maintenance"`
	Message          string `json:"message"`
	EstimatedUntil   string `json:"estimated_until,omitempty"`
	RemainingSeconds int64  `json:"remaining_seconds,omitempty"`
}

type MaintenanceStatusReader interface {
	Get() (*model.MaintenanceStatus, error)
}

func MaintenanceMode(repo MaintenanceStatusReader, jwtSvc *utils.JWTService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		status, err := repo.Get()
		if err != nil || !status.IsActive {
			return c.Next()
		}
		if isSuperAdmin(c, jwtSvc) {
			return c.Next()
		}

		msg := status.Message
		if msg == "" {
			msg = constant.MsgMaintenanceDefault
		}
		payload := MaintenancePayload{Maintenance: true, Message: msg}
		if status.EstimatedUntil != nil {
			payload.EstimatedUntil = status.EstimatedUntil.Format(time.RFC3339)
			if remaining := time.Until(*status.EstimatedUntil); remaining > 0 {
				payload.RemainingSeconds = int64(remaining.Seconds())
			}
		}
		return c.Status(fiber.StatusServiceUnavailable).JSON(payload)
	}
}

func isSuperAdmin(c *fiber.Ctx, jwtSvc *utils.JWTService) bool {
	header := c.Get("Authorization")
	if !strings.HasPrefix(header, "Bearer ") {
		return false
	}
	claims, err := jwtSvc.ParseAccessToken(strings.TrimPrefix(header, "Bearer "))
	if err != nil {
		return false
	}
	return claims.RoleName == constant.RoleSuperAdmin
}
