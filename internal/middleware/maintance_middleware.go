package middleware

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/projsonal/gostock/internal/model"
	"github.com/projsonal/gostock/pkg/constant"
	"github.com/projsonal/gostock/pkg/utils"
)

// MaintenancePayload adalah body respons 503 saat mode maintenance aktif —
// dibuat public (diekspor) supaya frontend punya bentuk data yang jelas
// untuk menampilkan banner + hitung mundur estimasi selesai.
type MaintenancePayload struct {
	Maintenance      bool   `json:"maintenance"`
	Message          string `json:"message"`
	EstimatedUntil   string `json:"estimated_until,omitempty"`
	RemainingSeconds int64  `json:"remaining_seconds,omitempty"`
}

// MaintenanceStatusReader adalah interface minimal (bukan mengimpor seluruh
// package repository maintenance) supaya middleware ini tidak terikat erat
// (tightly coupled) ke detail implementasi repository — sama seperti pola
// PermissionChecker pada rbac_middleware.go.
type MaintenanceStatusReader interface {
	Get() (*model.MaintenanceStatus, error)
}

// MaintenanceMode memblokir akses ke endpoint operasional selama mode
// maintenance aktif, KECUALI untuk pengguna dengan role super_admin (supaya
// admin tetap bisa login & mematikan mode maintenance tanpa akses database
// langsung). Status dibaca dari database di setiap request (lihat
// MaintenanceStatusReader) — bukan cache in-memory — supaya perubahan
// status oleh admin langsung berlaku di semua instance backend tanpa perlu
// restart.
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
