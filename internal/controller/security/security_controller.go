package security

import (
	"github.com/gofiber/fiber/v2"

	"github.com/projsonal/gostock/pkg/utils"
)

func (h *Controller) Check(c *fiber.Ctx) error {
	var req CheckRequest
	_ = c.BodyParser(&req) // body opsional, boleh kosong di percobaan pertama

	if req.BotToken != "" && h.botSvc.Verify(req.BotToken) {
		fresh, err := h.botSvc.Issue()
		if err != nil {
			return utils.Fail(c, fiber.StatusInternalServerError, "gagal menerbitkan token verifikasi", nil)
		}
		return utils.OK(c, "verifikasi otomatis berhasil", CheckResponse{Passed: true, BotToken: fresh})
	}

	challenge, err := h.captchaSvc.Generate()
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal membuat captcha", nil)
	}
	return utils.OK(c, "sesi tidak aktif terlalu lama, silakan selesaikan captcha", CheckResponse{
		Passed:  false,
		Captcha: challenge,
	})
}

func (h *Controller) Solve(c *fiber.Ctx) error {
	var req SolveRequest
	if !utils.ParseAndValidate(c, &req) {
		return nil
	}

	if err := h.captchaSvc.Verify(req.CaptchaToken, req.CaptchaAnswer); err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "verifikasi captcha gagal: "+err.Error(), nil)
	}

	token, err := h.botSvc.Issue()
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal menerbitkan token verifikasi", nil)
	}
	return utils.OK(c, "verifikasi berhasil", CheckResponse{Passed: true, BotToken: token})
}

func (h *Controller) RegisterRoutes(router fiber.Router) {
	g := router.Group("/security")
	g.Post("/verify", h.Check)
	g.Post("/challenge", h.Solve)
}
