package captcha

import (
	"github.com/gofiber/fiber/v2"

	"github.com/projsonal/gowms/pkg/utils"
)

// Generate GET /api/v1/captcha — mengembalikan captcha_token (dipakai lagi
// saat submit login) + captcha_image_base64 (PNG, tampilkan langsung di
// <img src="..."> pada frontend). TIDAK mengembalikan jawaban dalam bentuk
// apa pun — jawaban hanya bisa didapat dengan membaca gambar.
func (h *Controller) Generate(c *fiber.Ctx) error {
	challenge, err := h.svc.Generate()
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal membuat captcha", nil)
	}
	return utils.OK(c, "captcha berhasil dibuat", challenge)
}

// RegisterRoutes mendaftarkan endpoint publik untuk ambil captcha (harus
// bisa diakses sebelum login, jadi tanpa JWTAuth).
func (h *Controller) RegisterRoutes(router fiber.Router) {
	router.Get("/captcha", h.Generate)
}
