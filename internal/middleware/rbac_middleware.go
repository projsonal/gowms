package middleware

import (
	"github.com/gofiber/fiber/v2"

	"github.com/projsonal/gowms/pkg/constant"
	"github.com/projsonal/gowms/pkg/utils"
)

type PermissionChecker interface {
	HasPermission(roleID uint, module, action string) (bool, error)
}

func RequirePermission(checker PermissionChecker, module, action string) fiber.Handler {
	return func(c *fiber.Ctx) error {

		if roleName, _ := c.Locals(constant.CtxRoleName).(string); roleName == constant.RoleSuperAdmin {
			return c.Next()
		}

		roleID, ok := c.Locals(constant.CtxRoleID).(uint)
		if !ok {
			return utils.Fail(c, fiber.StatusUnauthorized, "sesi tidak valid", nil)
		}

		allowed, err := checker.HasPermission(roleID, module, action)
		if err != nil {
			return utils.Fail(c, fiber.StatusInternalServerError, "gagal memeriksa hak akses", nil)
		}
		if !allowed {
			return utils.Fail(c, fiber.StatusForbidden, "role anda tidak memiliki izin untuk aksi ini", nil)
		}
		return c.Next()
	}
}

func RequireRole(allowedRoles ...string) fiber.Handler {
	roleSet := make(map[string]struct{}, len(allowedRoles))
	for _, r := range allowedRoles {
		roleSet[r] = struct{}{}
	}

	return func(c *fiber.Ctx) error {
		roleName, _ := c.Locals(constant.CtxRoleName).(string)
		if _, ok := roleSet[roleName]; !ok {
			return utils.Fail(c, fiber.StatusForbidden, "role anda tidak diizinkan mengakses resource ini", nil)
		}
		return c.Next()
	}
}
