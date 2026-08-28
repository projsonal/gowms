package assettype

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/projsonal/gowms/internal/middleware"
	"github.com/projsonal/gowms/internal/model"
	"github.com/projsonal/gowms/pkg/constant"
	"github.com/projsonal/gowms/pkg/utils"
)

func normalizeKode(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func (h *Controller) List(c *fiber.Ctx) error {
	list, err := h.repo.List()
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal mengambil daftar jenis aset", nil)
	}
	return utils.OK(c, "daftar jenis aset berhasil diambil", list)
}

func (h *Controller) Create(c *fiber.Ctx) error {
	var req AssetTypeRequest
	if !utils.ParseAndValidate(c, &req) {
		return nil
	}
	kode := normalizeKode(req.Kode)
	if _, err := h.repo.FindByKode(kode); err == nil {
		return utils.Fail(c, fiber.StatusConflict, "kode jenis aset ini sudah dipakai", nil)
	}

	t := &model.AssetType{
		Kode:         kode,
		Label:        strings.TrimSpace(req.Label),
		Color:        req.Color,
		Abbr:         strings.ToUpper(strings.TrimSpace(req.Abbr)),
		HasKoordinat: *req.HasKoordinat,
		HasPort:      *req.HasPort,
		Urutan:       req.Urutan,
		IsSystem:     false,
	}
	if t.Color == "" {
		t.Color = "#6b7280"
	}
	if err := h.repo.Create(t); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal membuat jenis aset", nil)
	}
	return utils.Created(c, "jenis aset baru berhasil ditambahkan", t)
}

func (h *Controller) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "Data id jenis aset tidak valid", nil)
	}
	t, err := h.repo.FindByID(uint(id))
	if err != nil {
		return utils.Fail(c, fiber.StatusNotFound, "jenis aset tidak ditemukan", nil)
	}

	var req AssetTypeRequest
	if !utils.ParseAndValidate(c, &req) {
		return nil
	}
	kode := normalizeKode(req.Kode)
	if kode != t.Kode {
		if existing, ferr := h.repo.FindByKode(kode); ferr == nil && existing.ID != t.ID {
			return utils.Fail(c, fiber.StatusConflict, "kode jenis aset ini sudah dipakai", nil)
		}
		if t.IsSystem {
			return utils.Fail(c, fiber.StatusUnprocessableEntity, "kode jenis aset bawaan sistem tidak bisa diubah", nil)
		}
	}

	t.Kode = kode
	t.Label = strings.TrimSpace(req.Label)
	if req.Color != "" {
		t.Color = req.Color
	}
	t.Abbr = strings.ToUpper(strings.TrimSpace(req.Abbr))
	t.HasKoordinat = *req.HasKoordinat
	t.HasPort = *req.HasPort
	t.Urutan = req.Urutan

	if err := h.repo.Update(t); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal memperbarui jenis aset", nil)
	}
	return utils.OK(c, "jenis aset berhasil diperbarui", t)
}

func (h *Controller) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "Data id jenis aset tidak valid", nil)
	}
	if err := h.repo.Delete(uint(id)); err != nil {
		return utils.Fail(c, fiber.StatusUnprocessableEntity, err.Error(), nil)
	}
	return utils.OK(c, "jenis aset berhasil dihapus", nil)
}

func (h *Controller) RegisterRoutes(router fiber.Router) {
	g := router.Group("/asset-types", middleware.JWTAuth(h.jwtSvc))
	g.Get("/", h.List)
	g.Post("/", middleware.RequireRole(constant.RoleSuperAdmin, constant.RoleAdmin), h.Create)
	g.Put("/:id", middleware.RequireRole(constant.RoleSuperAdmin, constant.RoleAdmin), h.Update)
	g.Delete("/:id", middleware.RequireRole(constant.RoleSuperAdmin, constant.RoleAdmin), h.Delete)
}
