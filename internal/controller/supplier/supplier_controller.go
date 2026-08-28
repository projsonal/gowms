package supplier

import (
	"strconv"
	"strings"

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

func maskProtectedOne(role string, s *model.Supplier) {
	if role == constant.RoleSuperAdmin || role == constant.RoleAdmin || !s.IsProtected {
		return
	}
	locked := "*** dilindungi ***"
	s.Telepon = locked
	s.KerjasamaKurir = locked
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

func parseKurirNames(raw string) []string {
	parts := strings.Split(raw, ",")
	names := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			names = append(names, trimmed)
		}
	}
	return names
}

func (h *Controller) withStats(s model.Supplier) SupplierResponse {
	kurirNames := parseKurirNames(s.KerjasamaKurir)
	total, terkirim, err := h.repo.KurirStats(kurirNames)
	if err != nil {
		return SupplierResponse{Supplier: s}
	}
	var rating float64
	if total > 0 {
		rating = (float64(terkirim) / float64(total)) * 5
	}
	return SupplierResponse{Supplier: s, TotalOrder: total, Rating: rating}
}

func withStatsList(h *Controller, list []model.Supplier) []SupplierResponse {
	out := make([]SupplierResponse, 0, len(list))
	for _, s := range list {
		out = append(out, h.withStats(s))
	}
	return out
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
	return utils.OKWithMeta(c, "daftar supplier berhasil diambil", withStatsList(h, list), utils.BuildPaginationMeta(p, total))
}

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
	return utils.OK(c, "detail supplier berhasil diambil", h.withStats(*s))
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
		Kode:           req.Kode,
		Nama:           req.Nama,
		PIC:            req.PIC,
		Telepon:        req.Telepon,
		KerjasamaKurir: req.KerjasamaKurir,
		Alamat:         req.Alamat,
		NPWP:           req.NPWP,
		Catatan:        req.Catatan,
		IsActive:       true,
	}
	if err := h.repo.Create(s); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal membuat supplier", nil)
	}
	return utils.Created(c, "supplier berhasil dibuat", h.withStats(*s))
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
	s.KerjasamaKurir = req.KerjasamaKurir
	s.Alamat = req.Alamat
	s.NPWP = req.NPWP
	s.Catatan = req.Catatan
	if err := h.repo.Update(s); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal memperbarui supplier", nil)
	}
	return utils.OK(c, "supplier berhasil diperbarui", h.withStats(*s))
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

	inUse, err := h.repo.InUse(id)
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal memeriksa riwayat transaksi supplier", nil)
	}
	if inUse {
		return utils.Fail(c, fiber.StatusConflict,
			"supplier ini masih punya riwayat Purchase Order/Barang Masuk — tidak bisa dihapus. Nonaktifkan supplier ini saja bila sudah tidak dipakai.", nil)
	}
	if err := h.repo.Delete(id); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal menghapus supplier", nil)
	}
	return utils.OK(c, "supplier berhasil dihapus", nil)
}

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
	g.Patch("/:id/protect", onlySuperAdmin, h.Protect)
}
