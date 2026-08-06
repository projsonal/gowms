package barang

import (
	"strconv"

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

func parseListFilter(c *fiber.Ctx) barangRepo.Filter {
	kategoriID, _ := strconv.ParseUint(c.Query("kategori_id", "0"), 10, 64)
	satuanID, _ := strconv.ParseUint(c.Query("satuan_id", "0"), 10, 64)
	return barangRepo.Filter{
		KategoriID:  uint(kategoriID),
		SatuanID:    uint(satuanID),
		StokMenipis: c.QueryBool("stok_menipis", false),
		OnlyActive:  c.Query("status", "") == "aktif",
	}
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
// @Router       /stockrsd/barang [get]
func (h *Controller) List(c *fiber.Ctx) error {
	p := utils.PaginationFromContext(c)
	f := parseListFilter(c)

	list, total, err := h.repo.List(p, f)
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal mengambil daftar barang", nil)
	}
	return utils.OKWithMeta(c, "daftar barang berhasil diambil", list, utils.BuildPaginationMeta(p, total))
}

// Detail godoc
// @Summary      Detail barang
// @Tags         Barang
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "ID barang"
// @Success      200  {object}  utils.Envelope
// @Failure      404  {object}  utils.Envelope
// @Router       /stockrsd/barang/{id} [get]
func (h *Controller) Detail(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "id barang tidak valid", nil)
	}
	b, err := h.repo.FindByID(id)
	if err != nil {
		return utils.Fail(c, fiber.StatusNotFound, "barang tidak ditemukan", nil)
	}
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

// Create godoc
// @Summary      Tambah barang baru
// @Tags         Barang
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        payload  body      BarangRequest  true  "Data barang"
// @Success      201      {object}  utils.Envelope
// @Failure      400      {object}  utils.Envelope
// @Failure      409      {object}  utils.Envelope  "kode barang sudah dipakai"
// @Router       /stockrsd/barang [post]
func (h *Controller) Create(c *fiber.Ctx) error {
	var req BarangRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "payload tidak valid", nil)
	}
	if errs := utils.Validate(req); errs != nil {
		return utils.Fail(c, fiber.StatusUnprocessableEntity, "validasi gagal", errs)
	}

	if _, err := h.repo.FindByKode(req.KodeBarang); err == nil {
		return utils.Fail(c, fiber.StatusConflict, "kode barang sudah digunakan", nil)
	}
	if err := h.validateReferensi(req.KategoriID, req.SatuanID); err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "kategori atau satuan tidak ditemukan", nil)
	}

	b := &model.Barang{
		KodeBarang:  req.KodeBarang,
		Nama:        req.Nama,
		KategoriID:  req.KategoriID,
		SatuanID:    req.SatuanID,
		HargaBeli:   req.HargaBeli,
		StokMinimum: req.StokMinimum,
		IsActive:    true,
		Deskripsi:   req.Deskripsi,
	}
	if err := h.repo.Create(b); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal membuat barang", nil)
	}
	return utils.Created(c, "barang berhasil dibuat", b)
}

// Update PUT /barang/:id
func (h *Controller) Update(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "id barang tidak valid", nil)
	}
	b, err := h.repo.FindByID(id)
	if err != nil {
		return utils.Fail(c, fiber.StatusNotFound, "barang tidak ditemukan", nil)
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
		return utils.Fail(c, fiber.StatusBadRequest, "id barang tidak valid", nil)
	}
	b, err := h.repo.FindByID(id)
	if err != nil {
		return utils.Fail(c, fiber.StatusNotFound, "barang tidak ditemukan", nil)
	}
	if b.Stok > 0 {
		return utils.Fail(c, fiber.StatusConflict, "barang masih memiliki stok, kosongkan/pindahkan stok terlebih dahulu", nil)
	}

	if err := h.repo.Delete(id); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal menghapus barang", nil)
	}
	return utils.OK(c, "barang berhasil dihapus", nil)
}

func (h *Controller) UpdateStatus(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "id barang tidak valid", nil)
	}
	b, err := h.repo.FindByID(id)
	if err != nil {
		return utils.Fail(c, fiber.StatusNotFound, "barang tidak ditemukan", nil)
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

	g.Get("/summary", view, h.Summary)
	g.Get("/", view, h.List)
	g.Get("/:id", view, h.Detail)
	g.Post("/", tambah, h.Create)
	g.Put("/:id", edit, h.Update)
	g.Delete("/:id", edit, h.Delete)
	g.Patch("/:id/status", edit, h.UpdateStatus)
	g.Patch("/:id/adjust", edit, h.AdjustStok)
}
