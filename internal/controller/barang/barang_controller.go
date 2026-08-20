package barang

import (
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

// parseIDParam mengonversi parameter path ":id" ke uint dan memvalidasi
// formatnya, supaya path segmen non-numerik langsung dibalas 400 (bukan
// diam-diam jadi id=0 dan baru gagal belakangan di lookup DB).
func parseIDParam(c *fiber.Ctx) (uint, error) {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

// maskProtected menyamarkan field sensitif (harga beli & deskripsi) pada
// baris yang di-Protect, KHUSUS untuk role karyawan — supaya karyawan tetap
// bisa melihat baris itu ada di daftar (nama/SKU/stok) tapi tidak bisa
// mengecek data komersialnya. super_admin & admin tetap melihat data asli.
// Dipanggil di List/Detail sebelum data dikirim ke client (masking di sisi
// server, bukan sekadar disembunyikan di UI, supaya benar-benar tidak
// ikut terkirim lewat response API).
func maskProtectedOne(role string, b *model.Barang) {
	if role == constant.RoleSuperAdmin || role == constant.RoleAdmin || !b.IsProtected {
		return
	}
	b.HargaBeli = 0
	b.Deskripsi = "*** data dilindungi (Protect) — hubungi admin ***"
}

func maskProtected(role string, list []model.Barang) {
	for i := range list {
		maskProtectedOne(role, &list[i])
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

	// --- Visibilitas alur persetujuan (lihat model.Barang.ApprovalStatus) ---
	roleName, _ := c.Locals(constant.CtxRoleName).(string)
	userID, _ := c.Locals(constant.CtxUserID).(uint)
	switch roleName {
	case constant.RoleSuperAdmin:
		// super_admin lihat semua status secara default; ?approval_status=menunggu
		// dipakai halaman "Menunggu Persetujuan" untuk fokus ke antrean review.
		if s := c.Query("approval_status", ""); s != "" {
			f.ApprovalStatuses = []string{s}
		}
	case constant.RoleAdmin:
		// admin: lihat semua yang sudah disetujui, PLUS pengajuannya sendiri
		// apa pun statusnya (termasuk yang masih menunggu/ditolak).
		f.ApprovalStatuses = []string{constant.ApprovalDisetujui}
		f.OrSubmittedBy = userID
	default: // karyawan & role lain: HANYA yang sudah disetujui.
		f.ApprovalStatuses = []string{constant.ApprovalDisetujui}
	}
	return f
}

// List godoc
// @Summary      Daftar barang
// @Description  Daftar barang dengan pagination & filter.
// @Tags         Barang
// @Produce      json
// @Security     BearerAuth
// @Param        page          query     int     false  "Halaman"     default(1)
// @Param        limit         query     int     false  "Item per halaman"  default(10)
// @Param        search        query     string  false  "Kata kunci pencarian"
// @Param        kategori_id   query     int     false  "Filter kategori"
// @Param        satuan_id     query     int     false  "Filter satuan"
// @Param        stok_menipis  query     bool    false  "Hanya tampilkan stok menipis"
// @Param        status        query     string  false  "aktif untuk hanya barang aktif"
// @Success      200  {object}  utils.Envelope
// @Failure      401  {object}  utils.Envelope
// @Router       /wms-rsd/barang [get]
func (h *Controller) List(c *fiber.Ctx) error {
	p := utils.PaginationFromContext(c)
	f := parseListFilter(c)

	list, total, err := h.repo.List(p, f)
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal mengambil daftar barang", nil)
	}
	roleName, _ := c.Locals(constant.CtxRoleName).(string)
	maskProtected(roleName, list)
	return utils.OKWithMeta(c, "daftar barang berhasil diambil", list, utils.BuildPaginationMeta(p, total))
}

func (h *Controller) Detail(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "Data aset gudang `id barang` tidak valid", nil)
	}
	b, err := h.repo.FindByID(id)
	if err != nil {
		return utils.Fail(c, fiber.StatusNotFound, "Data Data barang tidak ditemukan", nil)
	}
	roleName, _ := c.Locals(constant.CtxRoleName).(string)
	userID, _ := c.Locals(constant.CtxUserID).(uint)
	if b.ApprovalStatus != constant.ApprovalDisetujui && roleName != constant.RoleSuperAdmin {
		isOwnSubmission := roleName == constant.RoleAdmin && b.DiajukanOleh != nil && *b.DiajukanOleh == userID
		if !isOwnSubmission {
			return utils.Fail(c, fiber.StatusNotFound, "Data Data barang tidak ditemukan", nil)
		}
	}
	maskProtectedOne(roleName, b)
	return utils.OK(c, "Data barang berhasil diambil", b)
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

func (h *Controller) Create(c *fiber.Ctx) error {
	var req BarangRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "proses payload tidak valid", nil)
	}
	if errs := utils.Validate(req); errs != nil {
		return utils.Fail(c, fiber.StatusUnprocessableEntity, " gagal memvalidasi", errs)
	}

	if _, err := h.repo.FindByKode(req.KodeBarang); err == nil {
		return utils.Fail(c, fiber.StatusConflict, "kode barang sudah digunakan", nil)
	}
	if err := h.validateReferensi(req.KategoriID, req.SatuanID); err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "kategori atau satuan tidak ditemukan", nil)
	}

	b := &model.Barang{
		KodeBarang:   req.KodeBarang,
		Nama:         req.Nama,
		KategoriID:   req.KategoriID,
		SatuanID:     req.SatuanID,
		HargaBeli:    req.HargaBeli,
		Stok:         req.Stok,
		StokMinimum:  req.StokMinimum,
		BeratGram:    req.BeratGram,
		IsActive:     true,
		IsSerialized: req.IsSerialized,
		Deskripsi:    req.Deskripsi,
	}

	// --- Alur persetujuan (lihat model.Barang.ApprovalStatus) ---
	// Barang yang dibuat admin TIDAK langsung aktif — menunggu Approve/Reject
	// dari super_admin dulu. super_admin membuat barang langsung disetujui
	// (nilai default kolom). Karyawan tetap memakai izin "tambah" biasa dari
	// matrix (lihat RegisterRoutes) — belum ada gerbang tambahan di sini.
	roleName, _ := c.Locals(constant.CtxRoleName).(string)
	if roleName == constant.RoleAdmin {
		userID, _ := c.Locals(constant.CtxUserID).(uint)
		b.ApprovalStatus = constant.ApprovalMenunggu
		b.DiajukanOleh = &userID
	}

	if err := h.repo.Create(b); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal membuat barang", nil)
	}
	msg := "barang berhasil dibuat"
	if b.ApprovalStatus == constant.ApprovalMenunggu {
		msg = "barang berhasil diajukan, menunggu persetujuan super admin"
	}
	return utils.Created(c, msg, b)
}

// Update PUT /barang/:id
func (h *Controller) Update(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "Data id barang tidak valid", nil)
	}
	b, err := h.repo.FindByID(id)
	if err != nil {
		return utils.Fail(c, fiber.StatusNotFound, "Data aset gudang tidak ditemukan", nil)
	}
	if b.IsProtected {
		return utils.Fail(c, fiber.StatusForbidden,
			"data ini dikunci (Protect) oleh super admin — buka kuncinya dulu sebelum diubah", nil)
	}

	var req BarangRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "payload tidak valid", nil)
	}
	if errs := utils.Validate(req); errs != nil {
		return utils.Fail(c, fiber.StatusUnprocessableEntity, "validasi gagal", errs)
	}

	if req.KodeBarang != b.KodeBarang {
		if existing, err := h.repo.FindByKode(req.KodeBarang); err == nil && existing.ID != b.ID {
			return utils.Fail(c, fiber.StatusConflict, "kode barang sudah digunakan", nil)
		}
	}
	if err := h.validateReferensi(req.KategoriID, req.SatuanID); err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "kategori atau satuan tidak ditemukan", nil)
	}

	b.KodeBarang = req.KodeBarang
	b.Nama = req.Nama
	b.KategoriID = req.KategoriID
	b.SatuanID = req.SatuanID
	b.HargaBeli = req.HargaBeli
	b.Stok = req.Stok
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
		return utils.Fail(c, fiber.StatusBadRequest, "Data id barang tidak valid", nil)
	}
	b, err := h.repo.FindByID(id)
	if err != nil {
		return utils.Fail(c, fiber.StatusNotFound, "Data barang tidak ditemukan", nil)
	}
	if b.IsProtected {
		return utils.Fail(c, fiber.StatusForbidden,
			"data ini dikunci (Protect) oleh super admin — buka kuncinya dulu sebelum dihapus", nil)
	}
	if b.Stok > 0 {
		return utils.Fail(c, fiber.StatusConflict, "barang masih memiliki stok, kosongkan/pindahkan stok terlebih dahulu", nil)
	}

	if err := h.repo.Delete(id); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal menghapus barang", nil)
	}
	return utils.OK(c, "barang berhasil dihapus", nil)
}

// Protect PATCH /barang/:id/protect — aksi "Protect" di action bar tabel.
// HANYA super_admin (lihat RegisterRoutes). Saat dikunci: barang tetap
// terlihat di daftar untuk semua role, tapi field sensitif (harga beli,
// deskripsi) disamarkan untuk karyawan, dan Update/Delete ditolak sampai
// dibuka kuncinya lagi.
func (h *Controller) Protect(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "id barang tidak valid", nil)
	}
	var req ProtectRequest
	if !utils.ParseAndValidate(c, &req) {
		return nil
	}
	b, err := h.repo.FindByID(id)
	if err != nil {
		return utils.Fail(c, fiber.StatusNotFound, "Data barang tidak ditemukan", nil)
	}
	b.IsProtected = *req.IsProtected
	if err := h.repo.Update(b); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal mengubah status proteksi", nil)
	}
	return utils.OK(c, "status proteksi berhasil diubah", b)
}

// Approve PATCH /barang/:id/approve — super_admin menyetujui pengajuan
// barang yang dibuat admin (lihat model.Barang.ApprovalStatus). Setelah
// disetujui, barang langsung tampil normal untuk semua role termasuk
// karyawan.
func (h *Controller) Approve(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "id barang tidak valid", nil)
	}
	b, err := h.repo.FindByID(id)
	if err != nil {
		return utils.Fail(c, fiber.StatusNotFound, "Data barang tidak ditemukan", nil)
	}
	if b.ApprovalStatus != constant.ApprovalMenunggu {
		return utils.Fail(c, fiber.StatusConflict, "barang ini tidak sedang menunggu persetujuan", nil)
	}
	userID, _ := c.Locals(constant.CtxUserID).(uint)
	now := time.Now()
	b.ApprovalStatus = constant.ApprovalDisetujui
	b.DisetujuiOleh = &userID
	b.DireviewPada = &now
	if err := h.repo.Update(b); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal menyetujui barang", nil)
	}
	return utils.OK(c, "barang berhasil disetujui", b)
}

// Reject PATCH /barang/:id/reject — super_admin menolak pengajuan barang
// dari admin, disertai catatan alasan penolakan.
func (h *Controller) Reject(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "id barang tidak valid", nil)
	}
	var req RejectRequest
	if !utils.ParseAndValidate(c, &req) {
		return nil
	}
	b, err := h.repo.FindByID(id)
	if err != nil {
		return utils.Fail(c, fiber.StatusNotFound, "Data barang tidak ditemukan", nil)
	}
	if b.ApprovalStatus != constant.ApprovalMenunggu {
		return utils.Fail(c, fiber.StatusConflict, "barang ini tidak sedang menunggu persetujuan", nil)
	}
	userID, _ := c.Locals(constant.CtxUserID).(uint)
	now := time.Now()
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
		return utils.Fail(c, fiber.StatusBadRequest, "id barang tidak valid", nil)
	}
	b, err := h.repo.FindByID(id)
	if err != nil {
		return utils.Fail(c, fiber.StatusNotFound, "Data barang tidak ditemukan", nil)
	}

	var req UpdateStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "payload tidak valid", nil)
	}
	if errs := utils.Validate(req); errs != nil {
		return utils.Fail(c, fiber.StatusUnprocessableEntity, "validasi gagal", errs)
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
		return utils.Fail(c, fiber.StatusBadRequest, "id barang tidak valid", nil)
	}

	var req AdjustStokRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "payload tidak valid", nil)
	}
	if errs := utils.Validate(req); errs != nil {
		return utils.Fail(c, fiber.StatusUnprocessableEntity, "validasi gagal", errs)
	}

	b, err := h.repo.AdjustStok(id, req.Delta)
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal memperbarui stok barang", nil)
	}
	return utils.OK(c, "stok barang berhasil diperbarui", b)
}

func (h *Controller) Summary(c *fiber.Ctx) error {
	total, err := h.repo.CountAll()
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal mengambil ringkasan", nil)
	}
	menipis, err := h.repo.CountStokMenipis()
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal mengambil ringkasan", nil)
	}
	nilai, err := h.repo.SumNilaiInventaris()
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal mengambil ringkasan", nil)
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
	// karyawan TIDAK PERNAH boleh menghapus data apa pun, meski matrix
	// izin "edit" untuknya diaktifkan untuk modul ini — Delete WAJIB staff.
	onlyStaff := middleware.RequireRole(constant.RoleSuperAdmin, constant.RoleAdmin)

	g.Get("/summary", view, h.Summary)
	g.Get("/", view, h.List)
	g.Get("/:id", view, h.Detail)
	g.Post("/", tambah, h.Create)
	g.Put("/:id", edit, h.Update)
	g.Delete("/:id", onlyStaff, edit, h.Delete)
	g.Patch("/:id/status", edit, h.UpdateStatus)
	g.Patch("/:id/adjust", edit, h.AdjustStok)
	g.Patch("/:id/protect", onlySuperAdmin, h.Protect) // Protect — khusus super admin
	g.Patch("/:id/approve", onlySuperAdmin, h.Approve) // Setujui pengajuan admin
	g.Patch("/:id/reject", onlySuperAdmin, h.Reject)   // Tolak pengajuan admin
}
