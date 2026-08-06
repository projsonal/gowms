package supplier

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	"github.com/projsonal/gowms/internal/middleware"
	"github.com/projsonal/gowms/internal/model"
	supplierRepo "github.com/projsonal/gowms/internal/repositories/supplier"
	"github.com/projsonal/gowms/pkg/constant"
	"github.com/projsonal/gowms/pkg/utils"
)

const Module = constant.ModuleSupplier

func parseIDParam(c *fiber.Ctx) (uint, error) {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

func (h *Controller) List(c *fiber.Ctx) error {
	p := utils.PaginationFromContext(c)
	f := supplierRepo.Filter{OnlyActive: c.Query("status", "") == "aktif"}

	list, total, err := h.repo.List(p, f)
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal mengambil daftar supplier", nil)
	}
	return utils.OKWithMeta(c, "daftar supplier berhasil diambil", list, utils.BuildPaginationMeta(p, total))
}

// Detail GET /supplier/:id
func (h *Controller) Detail(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "id supplier tidak valid", nil)
	}
	s, err := h.repo.FindByID(id)
	if err != nil {
		return utils.Fail(c, fiber.StatusNotFound, "supplier tidak ditemukan", nil)
	}
	return utils.OK(c, "detail supplier berhasil diambil", s)
}

func (h *Controller) Create(c *fiber.Ctx) error {
	var req SupplierRequest
	if !utils.ParseAndValidate(c, &req) {
		return nil
	}

	if _, err := h.repo.FindByKode(req.Kode); err == nil {
		return utils.Fail(c, fiber.StatusConflict, "kode supplier sudah digunakan", nil)
	}

	s := &model.Supplier{
		Kode:     req.Kode,
		Nama:     req.Nama,
		PIC:      req.PIC,
		Telepon:  req.Telepon,
		Email:    req.Email,
		Alamat:   req.Alamat,
		NPWP:     req.NPWP,
		Catatan:  req.Catatan,
		IsActive: true,
	}
	if err := h.repo.Create(s); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal membuat supplier", nil)
	}
	return utils.Created(c, "supplier berhasil dibuat", s)
}

func (h *Controller) Update(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "id supplier tidak valid", nil)
	}
	s, err := h.repo.FindByID(id)
	if err != nil {
		return utils.Fail(c, fiber.StatusNotFound, "supplier tidak ditemukan", nil)
	}

	var req SupplierRequest
	if !utils.ParseAndValidate(c, &req) {
		return nil
	}

	if req.Kode != s.Kode {
		if existing, err := h.repo.FindByKode(req.Kode); err == nil && existing.ID != s.ID {
			return utils.Fail(c, fiber.StatusConflict, "kode supplier sudah digunakan", nil)
		}
	}

	s.Kode = req.Kode
	s.Nama = req.Nama
	s.PIC = req.PIC
	s.Telepon = req.Telepon
	s.Email = req.Email
	s.Alamat = req.Alamat
	s.NPWP = req.NPWP
	s.Catatan = req.Catatan
	if err := h.repo.Update(s); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal memperbarui supplier", nil)
	}
	return utils.OK(c, "supplier berhasil diperbarui", s)
}

func (h *Controller) Delete(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "id supplier tidak valid", nil)
	}
	if _, err := h.repo.FindByID(id); err != nil {
		return utils.Fail(c, fiber.StatusNotFound, "supplier tidak ditemukan", nil)
	}
	if err := h.repo.Delete(id); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal menghapus supplier", nil)
	}
	return utils.OK(c, "supplier berhasil dihapus", nil)
}

func (h *Controller) UpdateStatus(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "id supplier tidak valid", nil)
	}
	s, err := h.repo.FindByID(id)
	if err != nil {
		return utils.Fail(c, fiber.StatusNotFound, "supplier tidak ditemukan", nil)
	}

	var req UpdateStatusRequest
	if !utils.ParseAndValidate(c, &req) {
		return nil
	}

	s.IsActive = *req.IsActive
	if err := h.repo.Update(s); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal memperbarui status supplier", nil)
	}
	return utils.OK(c, "status supplier berhasil diperbarui", s)
}

func (h *Controller) Summary(c *fiber.Ctx) error {
	total, err := h.repo.CountAll()
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal mengambil ringkasan", nil)
	}
	aktif, err := h.repo.CountActive()
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal mengambil ringkasan", nil)
	}

	return utils.OK(c, "ringkasan supplier berhasil diambil", SummaryResponse{
		TotalSupplier: total, SupplierAktif: aktif,
	})
}

func (h *Controller) RegisterRoutes(router fiber.Router) {
	g := router.Group("/supplier", middleware.JWTAuth(h.jwtSvc))

	view := middleware.RequirePermission(h.roleRepo, Module, constant.ActionView)
	tambah := middleware.RequirePermission(h.roleRepo, Module, constant.ActionTambah)
	edit := middleware.RequirePermission(h.roleRepo, Module, constant.ActionEdit)

	g.Get("/summary", view, h.Summary)
	g.Get("/", view, h.List)
	g.Get("/:id", view, h.Detail)
	g.Post("/", tambah, h.Create)
	g.Put("/:id", edit, h.Update)
	g.Delete("/:id", edit, h.Delete)
	g.Patch("/:id/status", edit, h.UpdateStatus)
}
