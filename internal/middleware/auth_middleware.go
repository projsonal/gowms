// Package middleware berisi seluruh middleware Fiber: autentikasi JWT dan
// otorisasi RBAC per-modul, dipakai oleh internal/routes/router.go.
package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/projsonal/gostock/pkg/constant"
	"github.com/projsonal/gostock/pkg/utils"
)

func JWTAuth(jwtSvc *utils.JWTService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		header := c.Get("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			return utils.Fail(c, fiber.StatusUnauthorized, "token tidak ditemukan", nil)
		}

		tokenStr := strings.TrimPrefix(header, "Bearer ")
		claims, err := jwtSvc.ParseAccessToken(tokenStr)
		if err != nil {
			return utils.Fail(c, fiber.StatusUnauthorized, "token tidak valid atau kedaluwarsa", nil)
		}

		c.Locals(constant.CtxUserID, claims.UserID)
		c.Locals(constant.CtxRoleID, claims.RoleID)
		c.Locals(constant.CtxRoleName, claims.RoleName)
		return c.Next()
	}
}
