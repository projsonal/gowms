package middleware

import (
	"log"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/projsonal/gostock/pkg/config"
)

// containsWildcard mengecek apakah salah satu origin di daftar adalah "*".
// Diekstrak jadi fungsi terpisah (bukan loop di dalam CORS()) supaya
// cognitive complexity CORS() tetap rendah — sejalan dengan aturan SonarQube
// yang membatasi jumlah percabangan bersarang dalam satu fungsi.
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
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, X-Bot-Token",
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
