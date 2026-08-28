package health

import (
	"github.com/gofiber/fiber/v2"

	"github.com/projsonal/gowms/pkg/utils"
)

func (h *Controller) Live(c *fiber.Ctx) error {
	return utils.OK(c, "aplikasi hidup", fiber.Map{"status": statusUp})
}

func (h *Controller) Ready(c *fiber.Ctx) error {
	return h.respondWithReport(c, "aplikasi siap menerima traffic", "aplikasi belum siap menerima traffic")
}

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
