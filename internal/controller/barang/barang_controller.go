package barang

import (
	"errors"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/projsonal/gowms/internal/middleware"
	"github.com/projsonal/gowms/internal/model"
	barangRepo "github.com/projsonal/gowms/internal/repositories/barang"
	"github.com/projsonal/gowms/pkg/constant"
	"github.com/projsonal/gowms/pkg/utils"
)

const Module = constant.ModuleKelolaBarang

const (
	msgIDBarangInvalid  = "id barang tidak valid"
	msgBarangNotFound   = "barang tidak ditemukan"
	msgReferensiInvalid = "kategori atau satuan tidak ditemukan"
)

func parseIDParam(c *fiber.Ctx) (uint, error) {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

func currentRole(c *fiber.Ctx) string {
	role, _ := c.Locals(constant.CtxRoleName).(string)
	return role
}

func currentUserID(c *fiber.Ctx) uint {
	id, _ := c.Locals(constant.CtxUserID).(uint)
	return id
}

// canViewUnapproved reports whether role can see a barang that isn't
// approved yet, either because they're super admin or because they
// submitted it themselves.
func canViewUnapproved(role string, userID uint, b *model.Barang) bool {
	if role == constant.RoleSuperAdmin {
		return true
	}
	return role == constant.RoleAdmin && b.DiajukanOleh != nil && *b.DiajukanOleh == userID
}

// maskProtectedOne hides commercially sensitive fields on a protected row
// for anyone below admin, without removing the row from the list.
func maskProtectedOne(role string, b *model.Barang) {
	if role == constant.RoleSuperAdmin || role == constant.RoleAdmin || !b.IsProtected {
		return
	}
	b.HargaBeli = 0
	b.Deskripsi = "*** data dilindungi (protect) — hubungi admin ***"
}

func maskProtected(role string, list []model.Barang) {
	for i := range list {
		maskProtectedOne(role, &list[i])
	}
}

// buildApprovalFilter scopes the barang list by approval status per role:
// super admin sees everything (optionally filtered), admin sees approved
// items plus their own submissions, everyone else sees approved only.
func buildApprovalFilter(c *fiber.Ctx, f *barangRepo.Filter) {
	role := currentRole(c)
	switch role {
	case constant.RoleSuperAdmin:
		if s := c.Query("approval_status", ""); s != "" {
			f.ApprovalStatuses = []string{s}
		}
	case constant.RoleAdmin:
		f.ApprovalStatuses = []string{constant.ApprovalDisetujui}
		f.OrSubmittedBy = currentUserID(c)
	default:
		f.ApprovalStatuses = []string{constant.ApprovalDisetujui}
	}
}

func parseListFilter(c *fiber.Ctx) barangRepo.Filter {
	kategoriID, _ := strconv.ParseUint(c.Query("kategori_id", "0"), 10, 64)
	satuanID, _ := strconv.ParseUint(c.Query("satuan_id", "0"), 10, 64)
	f := barangRepo.Filter{
		KategoriID:  uint(kategoriID),
		SatuanID:    uint(satuanID),
		StokMenipis: c.QueryBool("stok_menipis", false),
		OnlyActive:  c.Query("status", "") == "aktif",
	}
	buildApprovalFilter(c, &f)
	return f
}

func (h *Controller) List(c *fiber.Ctx) error {
	p := utils.PaginationFromContext(c)
	f := parseListFilter(c)

	list, total, err := h.repo.List(p, f)
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal mengambil daftar barang", nil)
	}
	maskProtected(currentRole(c), list)
	return utils.OKWithMeta(c, "daftar barang berhasil diambil", list, utils.BuildPaginationMeta(p, total))
}

func (h *Controller) Detail(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, msgIDBarangInvalid, nil)
	}
	b, err := h.repo.FindByID(id)
	if err != nil {
		return utils.Fail(c, fiber.StatusNotFound, msgBarangNotFound, nil)
	}
	role, userID := currentRole(c), currentUserID(c)
	if b.ApprovalStatus != constant.ApprovalDisetujui && !canViewUnapproved(role, userID, b) {
		return utils.Fail(c, fiber.StatusNotFound, msgBarangNotFound, nil)
	}
	maskProtectedOne(role, b)
	return utils.OK(c, "detail barang berhasil diambil", b)
}

func (h *Controller) validateReferensi(kategoriID, satuanID uint) error {
	if _, err := h.gudangRepo.FindKategoriByID(kategoriID); err != nil {
		return err
	}
	if _, err := h.gudangRepo.FindSatuanByID(satuanID); err != nil {
		return err
	}
	return nil
}

func toBarangModel(req BarangRequest) *model.Barang {
	return &model.Barang{
		KodeBarang:  req.KodeBarang,
		Nama:        req.Nama,
		KategoriID:  req.KategoriID,
		SatuanID:    req.SatuanID,
		HargaBeli:   req.HargaBeli,
		Stok:        req.StokAwal,
		StokMinimum: req.StokMinimum,
		BeratGram:   req.BeratGram,
		IsActive:    true,
		Deskripsi:   req.Deskripsi,
	}
}

func applySubmissionPolicy(c *fiber.Ctx, b *model.Barang) {
	if currentRole(c) != constant.RoleAdmin {
		return
	}
	userID := currentUserID(c)
	b.ApprovalStatus = constant.ApprovalMenunggu
	b.DiajukanOleh = &userID
}

func (h *Controller) Create(c *fiber.Ctx) error {
	var req BarangRequest
	if !utils.ParseAndValidate(c, &req) {
		return nil
	}
	if _, err := h.repo.FindByKode(req.KodeBarang); err == nil {
		return utils.Fail(c, fiber.StatusConflict, "kode barang sudah digunakan", nil)
	}
	if err := h.validateReferensi(req.KategoriID, req.SatuanID); err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, msgReferensiInvalid, nil)
	}

	b := toBarangModel(req)
	applySubmissionPolicy(c, b)

	if err := h.repo.Create(b); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal membuat barang", nil)
	}
	msg := "barang berhasil dibuat"
	if b.ApprovalStatus == constant.ApprovalMenunggu {
		msg = "barang berhasil diajukan, menunggu persetujuan super admin"
	}
	return utils.Created(c, msg, b)
}

func (h *Controller) Update(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, msgIDBarangInvalid, nil)
	}
	b, err := h.repo.FindByID(id)
	if err != nil {
		return utils.Fail(c, fiber.StatusNotFound, msgBarangNotFound, nil)
	}
	if b.IsProtected {
		return utils.Fail(c, fiber.StatusForbidden,
			"data yang dipil telah dikunci oleh super admin, hubungi admin untuk membuka kuncinya dulu sebelum ekskusi sesuai kebutuhan", nil)
	}

	var req BarangRequest
	if !utils.ParseAndValidate(c, &req) {
		return nil
	}
	if err := h.validateReferensi(req.KategoriID, req.SatuanID); err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, msgReferensiInvalid, nil)
	}

	b.KodeBarang = req.KodeBarang
	b.Nama = req.Nama
	b.KategoriID = req.KategoriID
	b.SatuanID = req.SatuanID
	b.HargaBeli = req.HargaBeli
	b.Stok = req.StokAwal
	b.BeratGram = req.BeratGram
	b.StokMinimum = req.StokMinimum
	b.Deskripsi = req.Deskripsi
	if err := h.repo.Update(b); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal memperbarui barang", nil)
	}
	return utils.OK(c, "barang berhasil diperbarui", b)
}

func (h *Controller) Delete(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, msgIDBarangInvalid, nil)
	}
	b, err := h.repo.FindByID(id)
	if err != nil {
		return utils.Fail(c, fiber.StatusNotFound, msgBarangNotFound, nil)
	}
	if b.IsProtected {
		return utils.Fail(c, fiber.StatusForbidden,
			"data yang dipil telah dikunci oleh super admin, hubungi admin untuk membuka kuncinya dulu sebelum ekskusi sesuai kebutuhan", nil)
	}
	if b.Stok > 0 {
		return utils.Fail(c, fiber.StatusConflict, "Data barang masih memiliki stok", nil)
	}
	if err := h.repo.Delete(id); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal menghapus barang", nil)
	}
	return utils.OK(c, "barang berhasil dihapus", nil)
}

func (h *Controller) Protect(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, msgIDBarangInvalid, nil)
	}
	var req ProtectRequest
	if !utils.ParseAndValidate(c, &req) {
		return nil
	}
	b, err := h.repo.FindByID(id)
	if err != nil {
		return utils.Fail(c, fiber.StatusNotFound, msgBarangNotFound, nil)
	}
	b.IsProtected = *req.IsProtected
	if err := h.repo.Update(b); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal mengubah status proteksi", nil)
	}
	return utils.OK(c, "proses proteksi terhadap data yang dipilih berhasil diubah", b)
}

func (h *Controller) requirePending(id uint) (*model.Barang, error) {
	b, err := h.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if b.ApprovalStatus != constant.ApprovalMenunggu {
		return nil, errors.New("barang ini tidak sedang menunggu persetujuan")
	}
	return b, nil
}

func (h *Controller) Approve(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, msgIDBarangInvalid, nil)
	}
	b, err := h.requirePending(id)
	if err != nil {
		return utils.Fail(c, fiber.StatusConflict, err.Error(), nil)
	}
	userID, now := currentUserID(c), time.Now()
	b.ApprovalStatus = constant.ApprovalDisetujui
	b.DisetujuiOleh = &userID
	b.DireviewPada = &now
	if err := h.repo.Update(b); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal menyetujui barang", nil)
	}
	return utils.OK(c, "barang berhasil disetujui", b)
}

func (h *Controller) Reject(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, msgIDBarangInvalid, nil)
	}
	var req RejectRequest
	if !utils.ParseAndValidate(c, &req) {
		return nil
	}
	b, err := h.requirePending(id)
	if err != nil {
		return utils.Fail(c, fiber.StatusConflict, err.Error(), nil)
	}
	userID, now := currentUserID(c), time.Now()
	b.ApprovalStatus = constant.ApprovalDitolak
	b.DisetujuiOleh = &userID
	b.CatatanApproval = req.Catatan
	b.DireviewPada = &now
	if err := h.repo.Update(b); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal menolak barang", nil)
	}
	return utils.OK(c, "barang berhasil ditolak", b)
}

func (h *Controller) UpdateStatus(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, msgIDBarangInvalid, nil)
	}
	b, err := h.repo.FindByID(id)
	if err != nil {
		return utils.Fail(c, fiber.StatusNotFound, msgBarangNotFound, nil)
	}
	var req UpdateStatusRequest
	if !utils.ParseAndValidate(c, &req) {
		return nil
	}
	b.IsActive = *req.IsActive
	if err := h.repo.Update(b); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal memperbarui status barang", nil)
	}
	return utils.OK(c, "status barang berhasil diperbarui", b)
}

func (h *Controller) AdjustStok(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, msgIDBarangInvalid, nil)
	}
	var req AdjustStokRequest
	if !utils.ParseAndValidate(c, &req) {
		return nil
	}
	b, err := h.repo.AdjustStok(id, req.Delta)
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal memperbarui stok barang", nil)
	}
	return utils.OK(c, "stok barang berhasil diperbarui", b)
}

func (h *Controller) Summary(c *fiber.Ctx) error {
	const msgGagalRingkasan = "gagal mengambil ringkasan"

	total, err := h.repo.CountAll()
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, msgGagalRingkasan, nil)
	}
	menipis, err := h.repo.CountStokMenipis()
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, msgGagalRingkasan, nil)
	}
	nilai, err := h.repo.SumNilaiInventaris()
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, msgGagalRingkasan, nil)
	}

	return utils.OK(c, "ringkasan barang berhasil diambil", SummaryResponse{
		TotalBarang: total, StokMenipis: menipis, TotalNilaiInventaris: nilai,
	})
}

func (h *Controller) RegisterRoutes(router fiber.Router) {
	g := router.Group("/barang", middleware.JWTAuth(h.jwtSvc))

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
	g.Patch("/:id/adjust", edit, h.AdjustStok)
	g.Patch("/:id/protect", onlySuperAdmin, h.Protect)
	g.Patch("/:id/approve", onlySuperAdmin, h.Approve)
	g.Patch("/:id/reject", onlySuperAdmin, h.Reject)
}
