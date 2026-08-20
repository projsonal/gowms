package barang_serial

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	"github.com/projsonal/gowms/internal/middleware"
	barangSerialRepo "github.com/projsonal/gowms/internal/repositories/barang_serial"
	"github.com/projsonal/gowms/pkg/constant"
	"github.com/projsonal/gowms/pkg/utils"
)

// Module — dipakai izin (RBAC) yang SAMA dengan modul Kelola Barang,
// karena unit/SN adalah rincian dari data barang itu sendiri, bukan
// modul tersendiri secara bisnis (lihat internal/controller/barang).
const Module = constant.ModuleKelolaBarang

func parseIDParam(c *fiber.Ctx) (uint, error) {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

// List GET /barang-serial?page=&limit=&search=&barang_id=&gudang_id=&status=
// search mencocokkan nomor_seri (ILIKE) — dipakai kotak pencarian umum di
// halaman daftar unit.
func (h *Controller) List(c *fiber.Ctx) error {
	p := utils.PaginationFromContext(c)
	barangID, _ := strconv.ParseUint(c.Query("barang_id", "0"), 10, 64)
	gudangID, _ := strconv.ParseUint(c.Query("gudang_id", "0"), 10, 64)
	f := barangSerialRepo.Filter{
		BarangID: uint(barangID),
		GudangID: uint(gudangID),
		Status:   c.Query("status", ""),
	}
	list, total, err := h.repo.List(p, f)
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal mengambil daftar unit barang", nil)
	}
	return utils.OKWithMeta(c, "daftar unit barang berhasil diambil", list, utils.BuildPaginationMeta(p, total))
}

// Create POST /barang-serial — pendaftaran unit MANUAL, khusus untuk
// mendigitalisasi stok fisik yang sudah ada di gudang SEBELUM modul
// pelacakan SN ini dipakai (padanan dengan field Stok saat Tambah Barang
// baru di internal/controller/barang). Menaikkan Barang.Stok +1 sekaligus
// — lihat barangSerialRepo.Repository.Create.
func (h *Controller) Create(c *fiber.Ctx) error {
	var req CreateRequest
	if !utils.ParseAndValidate(c, &req) {
		return nil
	}
	if _, err := h.barangRepo.FindByID(req.BarangID); err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, constant.ErrSerialBarangTidakAda, nil)
	}
	s, err := h.repo.Create(req.BarangID, req.GudangID, req.RakID, req.SerialNumber, req.Catatan)
	if err != nil {
		return utils.Fail(c, fiber.StatusConflict, err.Error(), nil)
	}
	return utils.Created(c, "unit berhasil didaftarkan, stok barang ikut bertambah", s)
}

// Detail GET /barang-serial/:id — sertakan nomor dokumen Barang
// Masuk/Keluar asal & tujuan (riwayat unit), lihat DetailResponse.
func (h *Controller) Detail(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "id unit tidak valid", nil)
	}
	s, err := h.repo.FindByID(id)
	if err != nil {
		return utils.Fail(c, fiber.StatusNotFound, "unit tidak ditemukan", nil)
	}
	nomorMasuk, nomorKeluar, err := h.repo.RiwayatDokumen(s)
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal mengambil riwayat unit", nil)
	}
	return utils.OK(c, "detail unit berhasil diambil", DetailResponse{
		BarangSerial: s, NomorBarangMasuk: nomorMasuk, NomorBarangKeluar: nomorKeluar,
	})
}

// Cari GET /barang-serial/cari/:sn — pencarian utama fitur pembeda
// barang fisik: scan/ketik satu SN untuk langsung tahu barang apa,
// status, lokasinya saat ini, DAN riwayat dokumen Masuk/Keluarnya —
// walau KodeBarang-nya sama dengan unit lain yang tampak identik di
// daftar Kelola Barang.
func (h *Controller) Cari(c *fiber.Ctx) error {
	sn := c.Params("sn")
	if sn == "" {
		return utils.Fail(c, fiber.StatusBadRequest, "nomor seri wajib diisi", nil)
	}
	s, err := h.repo.FindBySerial(sn)
	if err != nil {
		return utils.Fail(c, fiber.StatusNotFound, constant.ErrSerialTidakDitemukan, nil)
	}
	nomorMasuk, nomorKeluar, err := h.repo.RiwayatDokumen(s)
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal mengambil riwayat unit", nil)
	}
	return utils.OK(c, "unit ditemukan", DetailResponse{
		BarangSerial: s, NomorBarangMasuk: nomorMasuk, NomorBarangKeluar: nomorKeluar,
	})
}

// Ringkasan GET /barang-serial/ringkasan/:barang_id — hitungan unit per
// status untuk satu barang, dipakai kartu ringkas di halaman detail
// Kelola Barang.
func (h *Controller) Ringkasan(c *fiber.Ctx) error {
	barangID, err := strconv.ParseUint(c.Params("barang_id"), 10, 64)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "id barang tidak valid", nil)
	}
	tersedia, terpasang, rusak, err := h.repo.CountByBarang(uint(barangID))
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal mengambil ringkasan unit", nil)
	}
	return utils.OK(c, "ringkasan unit berhasil diambil", RingkasanResponse{
		BarangID: uint(barangID), Tersedia: tersedia, Terpasang: terpasang, Rusak: rusak,
	})
}

// UpdateStatus PATCH /barang-serial/:id/status — tandai rusak/tersedia
// secara manual, di luar alur dokumen Barang Masuk/Keluar.
func (h *Controller) UpdateStatus(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "id unit tidak valid", nil)
	}
	var req UpdateStatusRequest
	if !utils.ParseAndValidate(c, &req) {
		return nil
	}
	s, err := h.repo.UpdateStatusManual(id, req.Status, req.Catatan)
	if err != nil {
		return utils.Fail(c, fiber.StatusConflict, err.Error(), nil)
	}
	return utils.OK(c, "status unit berhasil diperbarui", s)
}

// Delete DELETE /barang-serial/:id — hapus baris unit yang salah input
// (mis. salah scan SN). Soft-delete, lihat model.BarangSerial.
func (h *Controller) Delete(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "id unit tidak valid", nil)
	}
	if err := h.repo.Delete(id); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal menghapus unit", nil)
	}
	return utils.OK(c, "unit berhasil dihapus", nil)
}

func (h *Controller) RegisterRoutes(router fiber.Router) {
	g := router.Group("/barang-serial", middleware.JWTAuth(h.jwtSvc))

	view := middleware.RequirePermission(h.roleRepo, Module, constant.ActionView)
	tambah := middleware.RequirePermission(h.roleRepo, Module, constant.ActionTambah)
	edit := middleware.RequirePermission(h.roleRepo, Module, constant.ActionEdit)
	onlyStaff := middleware.RequireRole(constant.RoleSuperAdmin, constant.RoleAdmin)

	g.Get("/", view, h.List)
	g.Post("/", tambah, h.Create)
	g.Get("/cari/:sn", view, h.Cari)
	g.Get("/ringkasan/:barang_id", view, h.Ringkasan)
	g.Get("/:id", view, h.Detail)
	g.Patch("/:id/status", edit, h.UpdateStatus)
	g.Delete("/:id", onlyStaff, edit, h.Delete)
}
