package barang_keluar

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"

	notification "github.com/projsonal/gowms/internal/controller/notification"
	"github.com/projsonal/gowms/internal/middleware"
	"github.com/projsonal/gowms/internal/model"
	bkRepo "github.com/projsonal/gowms/internal/repositories/barang_keluar"
	"github.com/projsonal/gowms/pkg/constant"
	"github.com/projsonal/gowms/pkg/utils"
)

const Module = constant.ModuleBarangKeluar

const msgId = "id barang keluar tidak valid"

func parseIDParam(c *fiber.Ctx) (uint, error) {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

func generateNomorBK() string {
	return fmt.Sprintf("BK-%d-%d", time.Now().Year(), time.Now().UnixNano()%100000)
}

// validateItems memastikan tiap barang_id & rak_id (bila diisi) memang ada.
// Validasi KECUKUPAN stok/rak sengaja TIDAK dilakukan di sini — itu
// dilakukan atomik di dalam transaksi repository saat Complete, supaya
// tidak ada celah waktu (TOCTOU) antara pengecekan dan pengurangan stok
// jika ada dua pengeluaran barang yang sama diproses hampir bersamaan.
func (h *Controller) validateItems(items []ItemRequest) error {
	for _, it := range items {
		if _, err := h.barangRepo.FindByID(it.BarangID); err != nil {
			return fmt.Errorf("barang id %d tidak ditemukan", it.BarangID)
		}
		if it.RakID != nil {
			if _, err := h.gudangRepo.FindRakByID(*it.RakID); err != nil {
				return fmt.Errorf("rak id %d tidak ditemukan", *it.RakID)
			}
		}
	}
	return nil
}

func toItemModels(items []ItemRequest) []model.BarangKeluarItem {
	out := make([]model.BarangKeluarItem, 0, len(items))
	for _, it := range items {
		out = append(out, model.BarangKeluarItem{BarangID: it.BarangID, RakID: it.RakID, Qty: it.Qty})
	}
	return out
}

func (h *Controller) List(c *fiber.Ctx) error {
	p := utils.PaginationFromContext(c)
	gudangID, _ := strconv.ParseUint(c.Query("gudang_id", "0"), 10, 64)
	kategoriID, _ := strconv.ParseUint(c.Query("kategori_id", "0"), 10, 64)
	f := bkRepo.Filter{Status: c.Query("status", ""), GudangID: uint(gudangID), KategoriID: uint(kategoriID)}

	list, total, err := h.repo.List(p, f)
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal mengambil daftar barang keluar", nil)
	}
	return utils.OKWithMeta(c, "daftar barang keluar berhasil diambil", list, utils.BuildPaginationMeta(p, total))
}

// Detail GET /barang-keluar/:id
func (h *Controller) Detail(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, msgId, nil)
	}
	bk, err := h.repo.FindByID(id)
	if err != nil {
		return utils.Fail(c, fiber.StatusNotFound, constant.ErrBKTidakDitemukan, nil)
	}
	return utils.OK(c, "detail barang keluar berhasil diambil", bk)
}

func (h *Controller) Create(c *fiber.Ctx) error {
	var req BKRequest
	if !utils.ParseAndValidate(c, &req) {
		return nil
	}
	tanggal, err := parseTanggalHarian(req.Tanggal)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "format tanggal tidak valid (YYYY-MM-DD)", nil)
	}
	if _, err := h.gudangRepo.FindGudangByID(req.GudangID); err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "gudang tidak ditemukan", nil)
	}
	if err := h.validateItems(req.Items); err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	bk := &model.BarangKeluar{
		NomorPengeluaran: generateNomorBK(),
		GudangID:         req.GudangID,
		Status:           constant.StatusBKDraft,
		Tanggal:          tanggal,
		Keperluan:        req.Keperluan,
		Penerima:         req.Penerima,
		Items:            toItemModels(req.Items),
	}
	if err := h.repo.Create(bk); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal membuat dokumen barang keluar", nil)
	}
	notification.Notify(h.notifRepo, "out",
		"Barang Keluar Baru",
		bk.NomorPengeluaran+" dicatat.",
		"/home/barang-keluar", nil, "all")
	return utils.Created(c, "dokumen barang keluar berhasil dibuat", bk)
}

func (h *Controller) requireDraft(id uint) (*model.BarangKeluar, error) {
	bk, err := h.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if bk.Status != constant.StatusBKDraft {
		return nil, errors.New(constant.ErrBKBukanDraft)
	}
	return bk, nil
}

func (h *Controller) Update(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "id barang keluar tidak valid", nil)
	}
	bk, err := h.requireDraft(id)
	if err != nil {
		return utils.Fail(c, fiber.StatusConflict, err.Error(), nil)
	}

	var req BKRequest
	if !utils.ParseAndValidate(c, &req) {
		return nil
	}
	tanggal, err := parseTanggalHarian(req.Tanggal)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "format tanggal tidak valid (YYYY-MM-DD)", nil)
	}
	if _, err := h.gudangRepo.FindGudangByID(req.GudangID); err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "gudang tidak ditemukan", nil)
	}
	if err := h.validateItems(req.Items); err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	bk.GudangID = req.GudangID
	bk.Tanggal = tanggal
	bk.Keperluan = req.Keperluan
	bk.Penerima = req.Penerima
	if err := h.repo.Update(bk, toItemModels(req.Items)); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal memperbarui dokumen barang keluar", nil)
	}
	return utils.OK(c, "dokumen barang keluar berhasil diperbarui", bk)
}

func (h *Controller) Delete(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "id barang keluar tidak valid", nil)
	}
	if _, err := h.requireDraft(id); err != nil {
		return utils.Fail(c, fiber.StatusConflict, err.Error(), nil)
	}
	if err := h.repo.Delete(id); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal menghapus dokumen barang keluar", nil)
	}
	return utils.OK(c, "dokumen barang keluar berhasil dihapus", nil)
}

func (h *Controller) Complete(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "id barang keluar tidak valid", nil)
	}
	userID, _ := c.Locals(constant.CtxUserID).(uint)

	bk, err := h.repo.Complete(id, userID)
	if err != nil {
		return utils.Fail(c, fiber.StatusConflict, err.Error(), nil)
	}
	h.notifyLowStock(bk.Items)
	return utils.OK(c, "barang keluar berhasil diselesaikan, stok & rak telah diperbarui", bk)
}

// notifyLowStock — kirim SATU notifikasi broadcast ("all") per barang yang
// stoknya BARU SAJA turun ke/di bawah stok_minimum akibat dokumen barang
// keluar ini. "Baru saja" dihitung dari stok SEBELUM dipotong (Stok saat
// ini + Qty yang baru dikeluarkan) dibandingkan stok SESUDAH — supaya
// TIDAK spam notifikasi berulang tiap ada barang keluar lain sementara
// barang itu memang sudah lama di bawah ambang (item.Barang di sini sudah
// mencerminkan Stok TERBARU, lihat repo.Complete -> FindByID di akhir).
// Preferensi "Peringatan Stok Minimum" ON/OFF di Settings -> Notifikasi
// SENGAJA tidak dicek di sini — itu preferensi TAMPILAN per user/device
// (lihat NotificationBell.tsx), bukan penentu apakah notifikasi ini boleh
// dibuat sama sekali (kalau tidak dibuat sama sekali, user yang preferensinya
// ON tidak akan pernah melihatnya juga).
func (h *Controller) notifyLowStock(items []model.BarangKeluarItem) {
	for _, item := range items {
		if item.Barang == nil || item.Barang.StokMinimum <= 0 {
			continue
		}
		stokSebelum := item.Barang.Stok + item.Qty
		justCrossed := stokSebelum > item.Barang.StokMinimum && item.Barang.Stok <= item.Barang.StokMinimum
		if !justCrossed {
			continue
		}
		notification.Notify(h.notifRepo, "stok_menipis",
			"Stok Menipis",
			item.Barang.Nama+" tersisa "+strconv.Itoa(item.Barang.Stok)+" (ambang minimum "+strconv.Itoa(item.Barang.StokMinimum)+").",
			"/home/kelola-barang", nil, "all")
	}
}

func (h *Controller) Batalkan(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "id barang keluar tidak valid", nil)
	}
	bk, err := h.repo.Batalkan(id)
	if err != nil {
		return utils.Fail(c, fiber.StatusConflict, err.Error(), nil)
	}
	return utils.OK(c, "dokumen barang keluar berhasil dibatalkan", bk)
}

// Summary GET /barang-keluar/summary
func (h *Controller) Summary(c *fiber.Ctx) error {
	total, err := h.repo.CountByStatus("")
	draft, err2 := h.repo.CountByStatus(constant.StatusBKDraft)
	selesai, err3 := h.repo.CountByStatus(constant.StatusBKSelesai)
	if err != nil || err2 != nil || err3 != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal mengambil ringkasan", nil)
	}
	return utils.OK(c, "ringkasan barang keluar berhasil diambil", SummaryResponse{
		TotalDokumen: total, Draft: draft, Selesai: selesai,
	})
}

// RegisterRoutes mendaftarkan endpoint modul "Barang Keluar".
func (h *Controller) RegisterRoutes(router fiber.Router) {
	g := router.Group("/barang-keluar", middleware.JWTAuth(h.jwtSvc))

	view := middleware.RequirePermission(h.roleRepo, Module, constant.ActionView)
	tambah := middleware.RequirePermission(h.roleRepo, Module, constant.ActionTambah)
	edit := middleware.RequirePermission(h.roleRepo, Module, constant.ActionEdit)
	onlyStaff := middleware.RequireRole(constant.RoleSuperAdmin, constant.RoleAdmin)

	g.Get("/summary", view, h.Summary)
	g.Get("/", view, h.List)
	g.Get("/:id", view, h.Detail)
	g.Post("/", tambah, h.Create)
	g.Put("/:id", edit, h.Update)
	g.Delete("/:id", onlyStaff, edit, h.Delete)
	g.Patch("/:id/selesai", edit, h.Complete)
	g.Patch("/:id/batalkan", edit, h.Batalkan)
}
