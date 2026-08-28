package barang

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	notification "github.com/projsonal/gowms/internal/controller/notifikasi"
	"github.com/projsonal/gowms/internal/middleware"
	"github.com/projsonal/gowms/internal/model"
	barangRepo "github.com/projsonal/gowms/internal/repositories/barang"
	"github.com/projsonal/gowms/internal/repositories/barangstokgudang"
	"github.com/projsonal/gowms/pkg/constant"
	"github.com/projsonal/gowms/pkg/utils"
)

const Module = constant.ModuleKelolaBarang

func parseIDParam(c *fiber.Ctx) (uint, error) {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

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

	roleName, _ := c.Locals(constant.CtxRoleName).(string)
	userID, _ := c.Locals(constant.CtxUserID).(uint)
	switch roleName {
	case constant.RoleSuperAdmin:

		if s := c.Query("approval_status", ""); s != "" {
			f.ApprovalStatuses = []string{s}
		}
	case constant.RoleAdmin:

		f.ApprovalStatuses = []string{constant.ApprovalDisetujui}
		f.OrSubmittedBy = userID
	case constant.RoleKaryawan:

		f.ApprovalStatuses = []string{constant.ApprovalDisetujui}
		f.OrSubmittedBy = userID
	default:
		f.ApprovalStatuses = []string{constant.ApprovalDisetujui}
	}
	return f
}

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
		isOwnSubmission := (roleName == constant.RoleAdmin || roleName == constant.RoleKaryawan) &&
			b.DiajukanOleh != nil && *b.DiajukanOleh == userID
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

	if req.Stok > 0 {
		if req.StokGudangID == 0 {
			return utils.Fail(c, fiber.StatusBadRequest,
				"stok awal diisi tapi gudang tujuannya belum dipilih — pilih gudang supaya stok ini tercatat di lokasi yang benar", nil)
		}
		if _, err := h.gudangRepo.FindGudangByID(req.StokGudangID); err != nil {
			return utils.Fail(c, fiber.StatusBadRequest, "gudang tujuan stok awal tidak ditemukan", nil)
		}
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
		Merek:        req.Merek,
		Tipe:         req.Tipe,
		Deskripsi:    req.Deskripsi,
	}

	roleName, _ := c.Locals(constant.CtxRoleName).(string)
	if roleName == constant.RoleAdmin || roleName == constant.RoleKaryawan {
		userID, _ := c.Locals(constant.CtxUserID).(uint)
		b.ApprovalStatus = constant.ApprovalMenunggu
		b.DiajukanOleh = &userID
	}

	if err := h.repo.Create(b); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal membuat barang", nil)
	}
	if req.Stok > 0 && req.StokGudangID != 0 {
		if err := h.repo.SetStokGudangAwal(b.ID, req.StokGudangID, req.Stok); err != nil {

			return utils.Fail(c, fiber.StatusInternalServerError,
				"barang berhasil dibuat, TAPI gagal mencatat stok awal ke gudang tujuan — buka Ubah Barang untuk mengisi ulang", nil)
		}
	}
	msg := "barang berhasil dibuat"
	if b.ApprovalStatus == constant.ApprovalMenunggu {
		msg = "barang berhasil diajukan, menunggu persetujuan super admin"
		notification.Notify(h.notifRepo, "barang",
			"Pengajuan Barang Baru",
			fmt.Sprintf("%s (%s) diajukan, menunggu pengecekan fisik & persetujuan.", b.Nama, b.KodeBarang),
			"/kelola-barang", nil, constant.RoleSuperAdmin)
	}
	return utils.Created(c, msg, b)
}

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

	b.IsSerialized = req.IsSerialized
	b.Merek = req.Merek
	b.Tipe = req.Tipe
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

func bisaMemprosesApproval(c *fiber.Ctx, b *model.Barang) bool {
	roleName, _ := c.Locals(constant.CtxRoleName).(string)
	if roleName == constant.RoleSuperAdmin {
		return true
	}
	userID, _ := c.Locals(constant.CtxUserID).(uint)
	return roleName == constant.RoleAdmin && b.DidelegasikanKe != nil && *b.DidelegasikanKe == userID
}

func (h *Controller) Approve(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "id barang tidak valid", nil)
	}
	b, err := h.repo.FindByID(id)
	if err != nil {
		return utils.Fail(c, fiber.StatusNotFound, "Data barang tidak ditemukan", nil)
	}
	if !bisaMemprosesApproval(c, b) {
		return utils.Fail(c, fiber.StatusForbidden, "kamu tidak berwenang memproses pengajuan ini — hanya super admin atau admin yang didelegasikan khusus baris ini", nil)
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
	if !bisaMemprosesApproval(c, b) {
		return utils.Fail(c, fiber.StatusForbidden, "kamu tidak berwenang memproses pengajuan ini — hanya super admin atau admin yang didelegasikan khusus baris ini", nil)
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

func (h *Controller) Delegasikan(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "id barang tidak valid", nil)
	}
	var req DelegasiRequest
	if !utils.ParseAndValidate(c, &req) {
		return nil
	}
	b, err := h.repo.FindByID(id)
	if err != nil {
		return utils.Fail(c, fiber.StatusNotFound, "Data barang tidak ditemukan", nil)
	}
	if b.ApprovalStatus != constant.ApprovalMenunggu {
		return utils.Fail(c, fiber.StatusConflict, "barang ini tidak sedang menunggu persetujuan — tidak perlu didelegasikan", nil)
	}
	target, err := h.usersRepo.FindByID(req.UserID)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "user tujuan delegasi tidak ditemukan", nil)
	}
	targetRole, err := h.roleRepo.FindByID(target.RoleID)
	if err != nil || targetRole.Name != constant.RoleAdmin {
		return utils.Fail(c, fiber.StatusUnprocessableEntity, "delegasi hanya bisa ke user dengan role admin", nil)
	}
	b.DidelegasikanKe = &req.UserID
	if err := h.repo.Update(b); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal mendelegasikan pengajuan", nil)
	}
	notification.Notify(h.notifRepo, "barang",
		"Pengajuan Barang Didelegasikan Kepadamu",
		fmt.Sprintf("%s (%s) menunggu pengecekan fisik & persetujuanmu.", b.Nama, b.KodeBarang),
		"/kelola-barang", &req.UserID, "")
	updated, _ := h.repo.FindByID(b.ID)
	return utils.OK(c, "pengajuan berhasil didelegasikan", updated)
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

func (h *Controller) RingkasanStok(c *fiber.Ctx) error {
	rows, err := barangstokgudang.ListAll(h.db)
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal mengambil ringkasan stok", nil)
	}
	out := make([]RingkasanStokRow, 0, len(rows))
	for _, r := range rows {
		out = append(out, RingkasanStokRow{
			BarangID: r.BarangID, KodeBarang: r.KodeBarang, NamaBarang: r.NamaBarang,
			GudangID: r.GudangID, NamaGudang: r.NamaGudang, Stok: r.Stok,
		})
	}
	return utils.OK(c, "ringkasan stok berhasil diambil", out)
}

func (h *Controller) NextSKU(c *fiber.Ctx) error {
	prefix := "BRG"
	if kategoriIDStr := c.Query("kategori_id"); kategoriIDStr != "" {
		kategoriID, err := strconv.ParseUint(kategoriIDStr, 10, 64)
		if err == nil {
			if kat, err := h.gudangRepo.FindKategoriByID(uint(kategoriID)); err == nil {
				prefix = derivePrefixFromName(kat.Nama)
			}
		}
	}
	if tipe := c.Query("tipe"); tipe != "" {
		if tipePrefix := derivePrefixFromName(tipe); tipePrefix != "BRG" {
			prefix = prefix + "-" + tipePrefix
		}
	}
	if merek := c.Query("merek"); merek != "" {
		if merekPrefix := derivePrefixFromName(merek); merekPrefix != "BRG" {
			prefix = prefix + "-" + merekPrefix
		}
	}
	if beratGramStr := c.Query("berat_gram"); beratGramStr != "" {
		if beratGram, err := strconv.Atoi(beratGramStr); err == nil {
			if ukuran := deriveUkuranFromBerat(&beratGram); ukuran != "" {
				prefix = prefix + "-" + ukuran
			}
		}
	}

	nomor, err := h.repo.NextSKUNumber(prefix)
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal membuat saran SKU", nil)
	}
	sku := fmt.Sprintf("%s-%04d", prefix, nomor)
	return utils.OK(c, "saran SKU berhasil dibuat", NextSKUResponse{SKU: sku})
}

func (h *Controller) CheckSKU(c *fiber.Ctx) error {
	sku := strings.TrimSpace(c.Query("sku"))
	if sku == "" {
		return utils.OK(c, "sku kosong", CheckSKUResponse{Available: false})
	}
	existing, err := h.repo.FindByKode(sku)
	if err != nil {
		return utils.OK(c, "sku tersedia", CheckSKUResponse{Available: true})
	}
	if editIDStr := c.Query("exclude_id"); editIDStr != "" {
		if editID, perr := strconv.ParseUint(editIDStr, 10, 64); perr == nil && existing.ID == uint(editID) {
			return utils.OK(c, "sku tersedia (milik barang yang sedang diedit)", CheckSKUResponse{Available: true})
		}
	}
	return utils.OK(c, "sku sudah dipakai", CheckSKUResponse{Available: false})
}

func derivePrefixFromName(nama string) string {
	var letters []rune
	for _, r := range strings.ToUpper(nama) {
		if r >= 'A' && r <= 'Z' {
			letters = append(letters, r)
			if len(letters) == 3 {
				break
			}
		}
	}
	if len(letters) == 0 {
		return "BRG"
	}
	return string(letters)
}

const (
	ukuranSmallMaxGram  = 1000
	ukuranMediumMaxGram = 5000
)

func deriveUkuranFromBerat(beratGram *int) string {
	if beratGram == nil {
		return ""
	}
	switch {
	case *beratGram < ukuranSmallMaxGram:
		return "S"
	case *beratGram <= ukuranMediumMaxGram:
		return "M"
	default:
		return "L"
	}
}

func (h *Controller) RegisterRoutes(router fiber.Router) {
	g := router.Group("/barang", middleware.JWTAuth(h.jwtSvc))

	view := middleware.RequirePermission(h.roleRepo, Module, constant.ActionView)
	tambah := middleware.RequirePermission(h.roleRepo, Module, constant.ActionTambah)
	edit := middleware.RequirePermission(h.roleRepo, Module, constant.ActionEdit)
	onlySuperAdmin := middleware.RequireRole(constant.RoleSuperAdmin)

	onlyStaff := middleware.RequireRole(constant.RoleSuperAdmin, constant.RoleAdmin)

	g.Get("/summary", view, h.Summary)
	g.Get("/ringkasan-stok", view, h.RingkasanStok)
	g.Get("/next-sku", view, h.NextSKU)
	g.Get("/check-sku", view, h.CheckSKU)
	g.Get("/", view, h.List)
	g.Get("/:id", view, h.Detail)
	g.Post("/", tambah, h.Create)
	g.Put("/:id", edit, h.Update)
	g.Delete("/:id", onlyStaff, edit, h.Delete)
	g.Patch("/:id/status", edit, h.UpdateStatus)
	g.Patch("/:id/adjust", edit, h.AdjustStok)
	g.Patch("/:id/protect", onlySuperAdmin, h.Protect)
	g.Patch("/:id/approve", onlyStaff, h.Approve)
	g.Patch("/:id/reject", onlyStaff, h.Reject)
	g.Patch("/:id/delegasikan", onlySuperAdmin, h.Delegasikan)
}
