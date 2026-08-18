package humancheck

import (
	"github.com/gofiber/fiber/v2"

	"github.com/projsonal/gowms/pkg/utils"
)

func (h *Controller) Issue(c *fiber.Ctx) error {
	token, err := h.svc.Issue()
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal membuat token verifikasi", nil)
	}
	return utils.OK(c, "token verifikasi berhasil dibuat", IssueResponse{Token: token})
}

func (h *Controller) RegisterRoutes(router fiber.Router) {
	router.Get("/human-check", h.Issue)
}
