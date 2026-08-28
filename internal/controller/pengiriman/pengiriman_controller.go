package pengiriman

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"

	notification "github.com/projsonal/gowms/internal/controller/notifikasi"
	"github.com/projsonal/gowms/internal/middleware"
	"github.com/projsonal/gowms/internal/model"
	pgRepo "github.com/projsonal/gowms/internal/repositories/pengiriman"
	"github.com/projsonal/gowms/pkg/constant"
	"github.com/projsonal/gowms/pkg/utils"
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
	tanggalKirim, err := parseTanggalHarian(req.TanggalKirim)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "format tanggal tidak valid (YYYY-MM-DD)", nil)
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
		DestLat:          req.DestLat,
		DestLng:          req.DestLng,
		Status:           constant.StatusPGDraft,
		TanggalKirim:     tanggalKirim,
		Catatan:          req.Catatan,
	}
	if err := h.repo.Create(pg); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal membuat dokumen pengiriman", nil)
	}
	notification.Notify(h.notifRepo, "ship",
		"Pengiriman Baru",
		pg.NomorPengiriman+" dibuat.",
		"/home/delivery", nil, "all")
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

func (h *Controller) Update(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "id pengiriman tidak valid", nil)
	}
	pg, err := h.requireDraft(id)
	if err != nil {
		return utils.Fail(c, fiber.StatusConflict, err.Error(), nil)
	}
	if pg.IsProtected {
		return utils.Fail(c, fiber.StatusForbidden,
			"data ini dikunci (Protect) oleh super admin — buka kuncinya dulu sebelum diubah", nil)
	}

	var req PengirimanRequest
	if !utils.ParseAndValidate(c, &req) {
		return nil
	}
	tanggalKirim, err := parseTanggalHarian(req.TanggalKirim)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "format tanggal tidak valid (YYYY-MM-DD)", nil)
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
	pg.DestLat = req.DestLat
	pg.DestLng = req.DestLng
	pg.TanggalKirim = tanggalKirim
	pg.Catatan = req.Catatan
	if err := h.repo.Update(pg); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal memperbarui dokumen pengiriman", nil)
	}
	return utils.OK(c, "dokumen pengiriman berhasil diperbarui", pg)
}

func (h *Controller) Delete(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "id pengiriman tidak valid", nil)
	}
	pg, err := h.requireDraft(id)
	if err != nil {
		return utils.Fail(c, fiber.StatusConflict, err.Error(), nil)
	}
	if pg.IsProtected {
		return utils.Fail(c, fiber.StatusForbidden,
			"data ini dikunci (Protect) oleh super admin — buka kuncinya dulu sebelum dihapus", nil)
	}
	if err := h.repo.Delete(id); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal menghapus dokumen pengiriman", nil)
	}
	return utils.OK(c, "dokumen pengiriman berhasil dihapus", nil)
}

func (h *Controller) ProtectPengiriman(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "id pengiriman tidak valid", nil)
	}
	var req ProtectRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "payload tidak valid", nil)
	}
	if errs := utils.Validate(req); errs != nil {
		return utils.Fail(c, fiber.StatusUnprocessableEntity, "validasi gagal", errs)
	}
	pg, err := h.repo.SetProtected(id, *req.IsProtected)
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal mengubah status proteksi", nil)
	}
	return utils.OK(c, "status proteksi berhasil diubah", pg)
}

func (h *Controller) Jadwalkan(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "id pengiriman tidak valid", nil)
	}
	var req JadwalkanRequest
	if !utils.ParseAndValidate(c, &req) {
		return nil
	}
	var estimasiTiba *time.Time
	if req.EstimasiTiba != "" {
		parsed, err := parseTanggalHarian(req.EstimasiTiba)
		if err != nil {
			return utils.Fail(c, fiber.StatusBadRequest, "format tanggal estimasi tidak valid (YYYY-MM-DD)", nil)
		}
		estimasiTiba = &parsed
	}

	pg, err := h.repo.Jadwalkan(id, req.NamaKurir, req.TeleponKurir, estimasiTiba)
	if err != nil {
		return utils.Fail(c, fiber.StatusConflict, err.Error(), nil)
	}
	return utils.OK(c, "pengiriman berhasil dijadwalkan", pg)
}

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

func (h *Controller) RegisterRoutes(router fiber.Router) {
	g := router.Group("/pengiriman", middleware.JWTAuth(h.jwtSvc))

	view := middleware.RequirePermission(h.roleRepo, Module, constant.ActionView)
	tambah := middleware.RequirePermission(h.roleRepo, Module, constant.ActionTambah)
	edit := middleware.RequirePermission(h.roleRepo, Module, constant.ActionEdit)
	onlyStaff := middleware.RequireRole(constant.RoleSuperAdmin, constant.RoleAdmin)
	onlySuperAdmin := middleware.RequireRole(constant.RoleSuperAdmin)

	g.Get("/summary", view, h.Summary)
	g.Get("/", view, h.List)
	g.Get("/:id", view, h.Detail)
	g.Get("/:id/lokasi", view, h.LokasiTerkini)
	g.Post("/", tambah, h.Create)
	g.Put("/:id", edit, h.Update)
	g.Delete("/:id", onlyStaff, edit, h.Delete)
	g.Patch("/:id/jadwalkan", edit, h.Jadwalkan)
	g.Patch("/:id/mulai", edit, h.Mulai)
	g.Post("/:id/lokasi", edit, h.KirimLokasi)
	g.Patch("/:id/selesai", edit, h.Selesaikan)
	g.Patch("/:id/batalkan", edit, h.Batalkan)
	g.Patch("/:id/protect", onlySuperAdmin, h.ProtectPengiriman)
}
