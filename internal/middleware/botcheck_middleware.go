package middleware

import (
	"github.com/gofiber/fiber/v2"

	"github.com/projsonal/gostock/pkg/botcheck"
	"github.com/projsonal/gostock/pkg/utils"
)

const BotTokenHeader = "X-Bot-Token"

type botTokenBody struct {
	BotToken string `json:"bot_token"`
}

func extraBotToken(c *fiber.Ctx) string {
	if token := c.Get(BotTokenHeader); token != "" {
		return token
	}
	var body botTokenBody
	_ = c.BodyParser(&body)
	return body.BotToken
}

func BotCheck(svc *botcheck.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := c.Get(BotTokenHeader)
		if token == "" || !svc.Verify(token) {
			return utils.Fail(c, fiber.StatusPreconditionRequired,
				"verifikasi bukan-bot diperlukan (sesi baru atau idle terlalu lama), selesaikan captcha lewat /security/challenge", nil)
		}

		fresh, err := svc.Issue()
		if err != nil {
			return utils.Fail(c, fiber.StatusInternalServerError, "gagal memperbarui sesi verifikasi", nil)
		}
		c.Set(BotTokenHeader, fresh)
		return c.Next()
	}
}
