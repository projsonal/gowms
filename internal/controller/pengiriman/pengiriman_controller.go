package pengiriman

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/projsonal/gostock/internal/middleware"
	"github.com/projsonal/gostock/internal/model"
	pgRepo "github.com/projsonal/gostock/internal/repositories/pengiriman"
	"github.com/projsonal/gostock/pkg/constant"
	"github.com/projsonal/gostock/pkg/utils"
)

const Module = constant.ModulePengiriman

const msgId = "id pengiriman tidak valid"

func parseIDParam(c *fiber.Ctx) (uint, error) {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

func generateNomorPG() string {
	return fmt.Sprintf("PG-%d-%d", time.Now().Year(), time.Now().UnixNano()%100000)
}

func (h *Controller) validateRequest(req PengirimanRequest) error {
	if _, err := h.gudangRepo.FindGudangByID(req.GudangAsalID); err != nil {
		return errors.New("gudang asal tidak ditemukan")
	}
	if req.JenisPengambilan == constant.JenisDropoff && req.AlamatTujuan == "" {
		return errors.New(constant.ErrPGAlamatWajib)
	}
	if req.BarangKeluarID != nil {
		if _, err := h.barangKeluarRepo.FindByID(*req.BarangKeluarID); err != nil {
			return errors.New("dokumen barang keluar tidak ditemukan")
		}
	}
	return nil
}

// List GET /pengiriman?page=&limit=&search=&status=&gudang_asal_id=&jenis=
func (h *Controller) List(c *fiber.Ctx) error {
	p := utils.PaginationFromContext(c)
	gudangID, _ := strconv.ParseUint(c.Query("gudang_asal_id", "0"), 10, 64)
	f := pgRepo.Filter{Status: c.Query("status", ""), GudangAsalID: uint(gudangID), Jenis: c.Query("jenis", "")}

	list, total, err := h.repo.List(p, f)
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal mengambil daftar pengiriman", nil)
	}
	return utils.OKWithMeta(c, "daftar pengiriman berhasil diambil", list, utils.BuildPaginationMeta(p, total))
}

// Detail GET /pengiriman/:id
func (h *Controller) Detail(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, msgId, nil)
	}
	pg, err := h.repo.FindByID(id)
	if err != nil {
		return utils.Fail(c, fiber.StatusNotFound, constant.ErrPGTidakDitemukan, nil)
	}
	return utils.OK(c, "detail pengiriman berhasil diambil", pg)
}

func (h *Controller) Create(c *fiber.Ctx) error {
	var req PengirimanRequest
	if !utils.ParseAndValidate(c, &req) {
		return nil
	}
	if err := h.validateRequest(req); err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	pg := &model.Pengiriman{
		NomorPengiriman:  generateNomorPG(),
		BarangKeluarID:   req.BarangKeluarID,
		GudangAsalID:     req.GudangAsalID,
		JenisPengambilan: req.JenisPengambilan,
		NamaPenerima:     req.NamaPenerima,
		TeleponPenerima:  req.TeleponPenerima,
		AlamatTujuan:     req.AlamatTujuan,
		Status:           constant.StatusPGDraft,
		TanggalKirim:     req.TanggalKirim,
		Catatan:          req.Catatan,
	}
	if err := h.repo.Create(pg); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal membuat dokumen pengiriman", nil)
	}
	return utils.Created(c, "dokumen pengiriman berhasil dibuat", pg)
}

func (h *Controller) requireDraft(id uint) (*model.Pengiriman, error) {
	pg, err := h.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if pg.Status != constant.StatusPGDraft {
		return nil, errors.New(constant.ErrPGBukanDraft)
	}
	return pg, nil
}

// Update PUT /pengiriman/:id — hanya boleh selama status masih draft.
func (h *Controller) Update(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "id pengiriman tidak valid", nil)
	}
	pg, err := h.requireDraft(id)
	if err != nil {
		return utils.Fail(c, fiber.StatusConflict, err.Error(), nil)
	}

	var req PengirimanRequest
	if !utils.ParseAndValidate(c, &req) {
		return nil
	}
	if err := h.validateRequest(req); err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	pg.BarangKeluarID = req.BarangKeluarID
	pg.GudangAsalID = req.GudangAsalID
	pg.JenisPengambilan = req.JenisPengambilan
	pg.NamaPenerima = req.NamaPenerima
	pg.TeleponPenerima = req.TeleponPenerima
	pg.AlamatTujuan = req.AlamatTujuan
	pg.TanggalKirim = req.TanggalKirim
	pg.Catatan = req.Catatan
	if err := h.repo.Update(pg); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal memperbarui dokumen pengiriman", nil)
	}
	return utils.OK(c, "dokumen pengiriman berhasil diperbarui", pg)
}

// Delete DELETE /pengiriman/:id — hanya boleh selama status draft.
func (h *Controller) Delete(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "id pengiriman tidak valid", nil)
	}
	if _, err := h.requireDraft(id); err != nil {
		return utils.Fail(c, fiber.StatusConflict, err.Error(), nil)
	}
	if err := h.repo.Delete(id); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal menghapus dokumen pengiriman", nil)
	}
	return utils.OK(c, "dokumen pengiriman berhasil dihapus", nil)
}

// Jadwalkan PATCH /pengiriman/:id/jadwalkan — draft -> dijadwalkan, assign kurir.
func (h *Controller) Jadwalkan(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "id pengiriman tidak valid", nil)
	}
	var req JadwalkanRequest
	if !utils.ParseAndValidate(c, &req) {
		return nil
	}

	pg, err := h.repo.Jadwalkan(id, req.NamaKurir, req.TeleponKurir, req.EstimasiTiba)
	if err != nil {
		return utils.Fail(c, fiber.StatusConflict, err.Error(), nil)
	}
	return utils.OK(c, "pengiriman berhasil dijadwalkan", pg)
}

// Mulai PATCH /pengiriman/:id/mulai — dijadwalkan -> dalam_perjalanan.
func (h *Controller) Mulai(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "id pengiriman tidak valid", nil)
	}
	pg, err := h.repo.Mulai(id)
	if err != nil {
		return utils.Fail(c, fiber.StatusConflict, err.Error(), nil)
	}
	return utils.OK(c, "pengiriman dimulai, silakan kirim update lokasi secara berkala", pg)
}

// KirimLokasi POST /pengiriman/:id/lokasi — ping posisi GPS kurir, hanya
// diterima selama status "dalam_perjalanan".
func (h *Controller) KirimLokasi(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "id pengiriman tidak valid", nil)
	}
	var req LokasiRequest
	if !utils.ParseAndValidate(c, &req) {
		return nil
	}

	recordedAt := time.Now()
	if req.RecordedAt != nil {
		recordedAt = *req.RecordedAt
	}

	pg, err := h.repo.RecordLocation(id, req.Lat, req.Lng, req.KecepatanKmh, recordedAt)
	if err != nil {
		return utils.Fail(c, fiber.StatusConflict, err.Error(), nil)
	}
	return utils.OK(c, "lokasi berhasil diperbarui", pg)
}

// LokasiTerkini GET /pengiriman/:id/lokasi?history=true&limit=200 —
// posisi terakhir kurir, dan (opsional) riwayat titik jalur untuk
// digambar sebagai rute di peta.
func (h *Controller) LokasiTerkini(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "id pengiriman tidak valid", nil)
	}
	pg, err := h.repo.FindByID(id)
	if err != nil {
		return utils.Fail(c, fiber.StatusNotFound, constant.ErrPGTidakDitemukan, nil)
	}

	type response struct {
		LastLat        *float64                        `json:"last_lat"`
		LastLng        *float64                        `json:"last_lng"`
		LastLocationAt *time.Time                      `json:"last_location_at"`
		Status         string                          `json:"status"`
		Riwayat        []model.PengirimanTrackingPoint `json:"riwayat,omitempty"`
	}
	res := response{LastLat: pg.LastLat, LastLng: pg.LastLng, LastLocationAt: pg.LastLocationAt, Status: pg.Status}

	if c.QueryBool("history", false) {
		limit := c.QueryInt("limit", 200)
		points, err := h.repo.ListTrackingPoints(id, limit)
		if err != nil {
			return utils.Fail(c, fiber.StatusInternalServerError, "gagal mengambil riwayat lokasi", nil)
		}
		res.Riwayat = points
	}
	return utils.OK(c, "lokasi pengiriman berhasil diambil", res)
}

// Selesaikan PATCH /pengiriman/:id/selesai — dalam_perjalanan -> terkirim.
func (h *Controller) Selesaikan(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "id pengiriman tidak valid", nil)
	}
	var req SelesaikanRequest
	if !utils.ParseAndValidate(c, &req) {
		return nil
	}

	pg, err := h.repo.Selesaikan(id, req.Catatan)
	if err != nil {
		return utils.Fail(c, fiber.StatusConflict, err.Error(), nil)
	}
	return utils.OK(c, "pengiriman berhasil diselesaikan", pg)
}

// Batalkan PATCH /pengiriman/:id/batalkan
func (h *Controller) Batalkan(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "id pengiriman tidak valid", nil)
	}
	pg, err := h.repo.Batalkan(id)
	if err != nil {
		return utils.Fail(c, fiber.StatusConflict, err.Error(), nil)
	}
	return utils.OK(c, "dokumen pengiriman berhasil dibatalkan", pg)
}

// Summary GET /pengiriman/summary
func (h *Controller) Summary(c *fiber.Ctx) error {
	total, err := h.repo.CountByStatus("")
	jalan, err2 := h.repo.CountByStatus(constant.StatusPGDalamPerjalanan)
	terkirim, err3 := h.repo.CountByStatus(constant.StatusPGTerkirim)
	if err != nil || err2 != nil || err3 != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal mengambil ringkasan", nil)
	}
	return utils.OK(c, "ringkasan pengiriman berhasil diambil", SummaryResponse{
		TotalPengiriman: total, DalamPerjalanan: jalan, Terkirim: terkirim,
	})
}

// RegisterRoutes mendaftarkan endpoint modul "Pengiriman".
func (h *Controller) RegisterRoutes(router fiber.Router) {
	g := router.Group("/pengiriman", middleware.JWTAuth(h.jwtSvc))

	view := middleware.RequirePermission(h.roleRepo, Module, constant.ActionView)
	tambah := middleware.RequirePermission(h.roleRepo, Module, constant.ActionTambah)
	edit := middleware.RequirePermission(h.roleRepo, Module, constant.ActionEdit)

	g.Get("/summary", view, h.Summary)
	g.Get("/", view, h.List)
	g.Get("/:id", view, h.Detail)
	g.Get("/:id/lokasi", view, h.LokasiTerkini)
	g.Post("/", tambah, h.Create)
	g.Put("/:id", edit, h.Update)
	g.Delete("/:id", edit, h.Delete)
	g.Patch("/:id/jadwalkan", edit, h.Jadwalkan)
	g.Patch("/:id/mulai", edit, h.Mulai)
	g.Post("/:id/lokasi", edit, h.KirimLokasi)
	g.Patch("/:id/selesai", edit, h.Selesaikan)
	g.Patch("/:id/batalkan", edit, h.Batalkan)
}
