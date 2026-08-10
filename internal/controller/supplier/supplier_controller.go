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

// maskProtectedOne menyamarkan field kontak sensitif (telepon, email,
// alamat, NPWP, catatan) pada supplier yang di-Protect, KHUSUS untuk role
// karyawan — baris tetap terlihat ada di daftar (nama, status) tapi
// datanya tidak bisa dicek. Masking dilakukan di server sebelum data
// dikirim, bukan cuma disembunyikan di UI.
func maskProtectedOne(role string, s *model.Supplier) {
	if role == constant.RoleSuperAdmin || role == constant.RoleAdmin || !s.IsProtected {
		return
	}
	locked := "*** dilindungi ***"
	s.Telepon = locked
	s.Email = locked
	s.Alamat = locked
	s.Catatan = locked
	if s.NPWP != nil {
		s.NPWP = &locked
	}
	if s.PIC != nil {
		s.PIC = &locked
	}
}

func maskProtected(role string, list []model.Supplier) {
	for i := range list {
		maskProtectedOne(role, &list[i])
	}
}

func (h *Controller) List(c *fiber.Ctx) error {
	p := utils.PaginationFromContext(c)
	f := supplierRepo.Filter{OnlyActive: c.Query("status", "") == "aktif"}

	list, total, err := h.repo.List(p, f)
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal mengambil daftar supplier", nil)
	}
	roleName, _ := c.Locals(constant.CtxRoleName).(string)
	maskProtected(roleName, list)
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
	roleName, _ := c.Locals(constant.CtxRoleName).(string)
	maskProtectedOne(roleName, s)
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
	if s.IsProtected {
		return utils.Fail(c, fiber.StatusForbidden,
			"data ini dikunci (Protect) oleh super admin — buka kuncinya dulu sebelum diubah", nil)
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
	s, err := h.repo.FindByID(id)
	if err != nil {
		return utils.Fail(c, fiber.StatusNotFound, "supplier tidak ditemukan", nil)
	}
	if s.IsProtected {
		return utils.Fail(c, fiber.StatusForbidden,
			"data ini dikunci (Protect) oleh super admin — buka kuncinya dulu sebelum dihapus", nil)
	}
	if err := h.repo.Delete(id); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal menghapus supplier", nil)
	}
	return utils.OK(c, "supplier berhasil dihapus", nil)
}

// Protect PATCH /supplier/:id/protect — aksi "Protect" di action bar
// tabel. HANYA super_admin (lihat RegisterRoutes). Sama pola dengan Barang/COD.
func (h *Controller) Protect(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "id supplier tidak valid", nil)
	}
	var req ProtectRequest
	if !utils.ParseAndValidate(c, &req) {
		return nil
	}
	s, err := h.repo.FindByID(id)
	if err != nil {
		return utils.Fail(c, fiber.StatusNotFound, "supplier tidak ditemukan", nil)
	}
	s.IsProtected = *req.IsProtected
	if err := h.repo.Update(s); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal mengubah status proteksi", nil)
	}
	return utils.OK(c, "status proteksi berhasil diubah", s)
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
	onlySuperAdmin := middleware.RequireRole(constant.RoleSuperAdmin)
	onlyStaff := middleware.RequireRole(constant.RoleSuperAdmin, constant.RoleAdmin)

	g.Get("/summary", view, h.Summary)
	g.Get("/", view, h.List)
	g.Get("/:id", view, h.Detail)
	g.Post("/", tambah, h.Create)
	g.Put("/:id", edit, h.Update)
	g.Delete("/:id", onlyStaff, edit, h.Delete)
	g.Patch("/:id/status", edit, h.UpdateStatus)
	g.Patch("/:id/protect", onlySuperAdmin, h.Protect) // Protect — khusus super admin
}
