package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/basicauth"
)

// SwaggerBasicAuth membungkus /swagger/* dengan HTTP Basic Auth sederhana.
// Dipakai supaya dokumentasi API tidak diakses bebas oleh publik kalau
// Swagger tetap diaktifkan di production (lihat SWAGGER_ENABLED,
// SWAGGER_BASIC_AUTH_USER, SWAGGER_BASIC_AUTH_PASS di pkg/config).
func SwaggerBasicAuth(user, pass string) fiber.Handler {
	return basicauth.New(basicauth.Config{
		Users: map[string]string{
			user: pass,
		},
		Realm: "GoWMS API Docs",
	})
}
