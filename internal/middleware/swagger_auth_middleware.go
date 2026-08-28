package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/basicauth"
)

func SwaggerBasicAuth(user, pass string) fiber.Handler {
	return basicauth.New(basicauth.Config{
		Users: map[string]string{
			user: pass,
		},
		Realm: "GoWMS API Docs",
	})
}
