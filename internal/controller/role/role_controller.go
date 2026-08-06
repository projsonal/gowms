package role

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	"github.com/projsonal/gostock/internal/middleware"
	"github.com/projsonal/gostock/internal/model"
	"github.com/projsonal/gostock/pkg/constant"
	"github.com/projsonal/gostock/pkg/utils"
)

func (h *Controller) List(c *fiber.Ctx) error {
	roles, err := h.repo.FindAll()
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal mengambil daftar role", nil)
	}
	return utils.OK(c, "daftar role berhasil diambil", roles)
}

func (h *Controller) Create(c *fiber.Ctx) error {
	var req CreateRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "payload tidak valid", nil)
	}
	if errs := utils.Validate(req); errs != nil {
		return utils.Fail(c, fiber.StatusUnprocessableEntity, "validasi gagal", errs)
	}

	if _, err := h.repo.FindByName(req.Name); err == nil {
		return utils.Fail(c, fiber.StatusConflict, "nama role sudah digunakan", nil)
	}

	roleModel := &model.Role{Name: req.Name, Description: req.Description}
	if err := h.repo.Create(roleModel); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal membuat role", nil)
	}
	return utils.Created(c, "role berhasil dibuat", roleModel)
}

func (h *Controller) GetPermissionMatrix(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "id role tidak valid", nil)
	}

	matrix, err := h.repo.GetMatrix(uint(id))
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal mengambil matrix akses", nil)
	}
	return utils.OK(c, "matrix akses berhasil diambil", fiber.Map{"role_id": id, "items": matrix})
}

func (h *Controller) UpdatePermissionMatrix(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "id role tidak valid", nil)
	}

	var req UpdatePermissionMatrixRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "payload tidak valid", nil)
	}
	if errs := utils.Validate(req); errs != nil {
		return utils.Fail(c, fiber.StatusUnprocessableEntity, "validasi gagal", errs)
	}

	var permissionIDs []uint
	for _, item := range req.Items {
		actions := []struct {
			enabled bool
			action  string
		}{
			{item.View, constant.ActionView},
			{item.Tambah, constant.ActionTambah},
			{item.Edit, constant.ActionEdit},
			{item.ApprovalReject, constant.ActionApprovalReject},
			{item.Print, constant.ActionPrint},
			{item.AssignDelegasi, constant.ActionAssignDelegasi},
		}

		for _, a := range actions {
			if !a.enabled {
				continue
			}
			p, err := h.repo.FindOrCreatePermission(item.Module, a.action)
			if err != nil {
				return utils.Fail(c, fiber.StatusInternalServerError, "gagal menyimpan matrix akses", nil)
			}
			permissionIDs = append(permissionIDs, p.ID)
		}
	}

	if err := h.repo.ReplaceRolePermissions(uint(id), permissionIDs); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal menyimpan matrix akses", nil)
	}
	return utils.OK(c, "matrix akses berhasil diperbarui", nil)
}

func (h *Controller) RegisterRoutes(router fiber.Router) {
	g := router.Group("/roles", middleware.JWTAuth(h.jwtSvc))
	g.Get("/", h.List)
	g.Post("/", middleware.RequireRole(constant.RoleSuperAdmin), h.Create)
	g.Get("/:id/permissions", h.GetPermissionMatrix)
	g.Put("/:id/permissions", middleware.RequireRole(constant.RoleSuperAdmin), h.UpdatePermissionMatrix)
}
