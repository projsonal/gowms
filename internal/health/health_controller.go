package health

import (
	"github.com/gofiber/fiber/v2"

	"github.com/projsonal/gowms/pkg/utils"
)

// Live godoc
// @Summary      Liveness probe
// @Description  Mengecek apakah proses aplikasi hidup (tidak mengecek dependensi eksternal seperti DB).
// @Tags         Health
// @Produce      json
// @Success      200  {object}  utils.Envelope
// @Router       /health/live [get]

func (h *Controller) Live(c *fiber.Ctx) error {
	return utils.OK(c, "aplikasi hidup", fiber.Map{"status": statusUp})
}

// Ready godoc
// @Summary      Readiness probe
// @Description  Mengecek apakah aplikasi siap menerima traffic (termasuk cek koneksi DB, dsb).
// @Tags         Health
// @Produce      json
// @Success      200  {object}  utils.Envelope
// @Failure      503  {object}  utils.Envelope
// @Router       /health/ready [get]
func (h *Controller) Ready(c *fiber.Ctx) error {
	return h.respondWithReport(c, "aplikasi siap menerima traffic", "aplikasi belum siap menerima traffic")
}

// Health godoc
// @Summary      Health check lengkap
// @Description  Laporan status seluruh komponen inti (DB, dsb).
// @Tags         Health
// @Produce      json
// @Success      200  {object}  utils.Envelope
// @Failure      503  {object}  utils.Envelope
// @Router       /health [get]
func (h *Controller) Health(c *fiber.Ctx) error {
	return h.respondWithReport(c, "seluruh komponen inti sehat", "salah satu komponen inti bermasalah")
}

func (h *Controller) respondWithReport(c *fiber.Ctx, upMessage, downMessage string) error {
	report := h.checker.Report()

	if report.Status == statusDown {
		return c.Status(fiber.StatusServiceUnavailable).JSON(utils.Envelope{
			Success: false,
			Message: downMessage,
			Data:    report,
		})
	}
	return utils.OK(c, upMessage, report)
}
