package middleware

import (
	"log"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/projsonal/gowms/pkg/config"
)

func containsWildcard(origins []string) bool {
	for _, o := range origins {
		if o == "*" {
			return true
		}
	}
	return false
}

func CORS(cfg *config.Config) fiber.Handler {
	origins := cfg.CORS.AllowedOrigins
	allowCredentials := !containsWildcard(origins)
	if !allowCredentials {
		log.Println("middleware: CORS_ALLOWED_ORIGINS=\"*\" terdeteksi — AllowCredentials dimatikan otomatis.")
	}

	return cors.New(cors.Config{
		AllowOrigins:     strings.Join(origins, ","),
		AllowMethods:     "GET, POST, PUT, PATCH, DELETE, OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, X-Bot-Token, X-Timezone",
		ExposeHeaders:    "Content-Length, X-Bot-Token",
		AllowCredentials: allowCredentials,
		MaxAge:           3600,
	})
}

func RequestLogger() fiber.Handler {
	return logger.New(logger.Config{
		Format:     "[${time}] ${status} - ${latency} ${method} ${path}\n",
		TimeFormat: time.RFC3339,
	})
}

func Recover() fiber.Handler {
	return recover.New()
}

func SecurityHeaders() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-Frame-Options", "DENY")
		c.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=(), payment=()")
		c.Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		c.Set("X-XSS-Protection", "0")
		if c.Protocol() == "https" {
			c.Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
		}
		return c.Next()
	}
}
