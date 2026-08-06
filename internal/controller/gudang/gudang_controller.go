package gudang

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	"github.com/projsonal/gostock/internal/middleware"
	"github.com/projsonal/gostock/internal/model"
	"github.com/projsonal/gostock/pkg/constant"
	"github.com/projsonal/gostock/pkg/utils"
)

const Module = constant.ModuleManajemenGudang

func parseIDParam(c *fiber.Ctx) (uint, error) {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

// ---- Kategori ----

// ListKategori GET /api/v1/gudang/kategori?page=&limit=&search=
func (h *Controller) ListKategori(c *fiber.Ctx) error {
	p := utils.PaginationFromContext(c)
	list, total, err := h.repo.ListKategori(p)
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal mengambil daftar kategori", nil)
	}
	return utils.OKWithMeta(c, "daftar kategori berhasil diambil", list, utils.BuildPaginationMeta(p, total))
}

// CreateKategori POST /api/v1/gudang/kategori
func (h *Controller) CreateKategori(c *fiber.Ctx) error {
	var req KategoriRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "payload tidak valid", nil)
	}
	if errs := utils.Validate(req); errs != nil {
		return utils.Fail(c, fiber.StatusUnprocessableEntity, "validasi gagal", errs)
	}
	if _, err := h.repo.FindKategoriByNama(req.Nama); err == nil {
		return utils.Fail(c, fiber.StatusConflict, "nama kategori sudah digunakan", nil)
	}

	k := &model.Kategori{Nama: req.Nama}
	if err := h.repo.CreateKategori(k); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal membuat kategori", nil)
	}
	return utils.Created(c, "kategori berhasil dibuat", k)
}

// UpdateKategori PUT /api/v1/gudang/kategori/:id
func (h *Controller) UpdateKategori(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "id kategori tidak valid", nil)
	}
	k, err := h.repo.FindKategoriByID(id)
	if err != nil {
		return utils.Fail(c, fiber.StatusNotFound, "kategori tidak ditemukan", nil)
	}

	var req KategoriRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "payload tidak valid", nil)
	}
	if errs := utils.Validate(req); errs != nil {
		return utils.Fail(c, fiber.StatusUnprocessableEntity, "validasi gagal", errs)
	}
	if req.Nama != k.Nama {
		if existing, err := h.repo.FindKategoriByNama(req.Nama); err == nil && existing.ID != k.ID {
			return utils.Fail(c, fiber.StatusConflict, "nama kategori sudah digunakan", nil)
		}
	}

	k.Nama = req.Nama
	if err := h.repo.UpdateKategori(k); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal memperbarui kategori", nil)
	}
	return utils.OK(c, "kategori berhasil diperbarui", k)
}

// DeleteKategori DELETE /api/v1/gudang/kategori/:id
func (h *Controller) DeleteKategori(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "id kategori tidak valid", nil)
	}
	if _, err := h.repo.FindKategoriByID(id); err != nil {
		return utils.Fail(c, fiber.StatusNotFound, "kategori tidak ditemukan", nil)
	}
	if err := h.repo.DeleteKategori(id); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal menghapus kategori", nil)
	}
	return utils.OK(c, "kategori berhasil dihapus", nil)
}

// ---- Satuan ----

// ListSatuan GET /api/v1/gudang/satuan?page=&limit=&search=
func (h *Controller) ListSatuan(c *fiber.Ctx) error {
	p := utils.PaginationFromContext(c)
	list, total, err := h.repo.ListSatuan(p)
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal mengambil daftar satuan", nil)
	}
	return utils.OKWithMeta(c, "daftar satuan berhasil diambil", list, utils.BuildPaginationMeta(p, total))
}

// CreateSatuan POST /api/v1/gudang/satuan
func (h *Controller) CreateSatuan(c *fiber.Ctx) error {
	var req SatuanRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "payload tidak valid", nil)
	}
	if errs := utils.Validate(req); errs != nil {
		return utils.Fail(c, fiber.StatusUnprocessableEntity, "validasi gagal", errs)
	}
	if _, err := h.repo.FindSatuanByNama(req.Nama); err == nil {
		return utils.Fail(c, fiber.StatusConflict, "nama satuan sudah digunakan", nil)
	}

	s := &model.Satuan{Nama: req.Nama, Singkatan: req.Singkatan}
	if err := h.repo.CreateSatuan(s); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal membuat satuan", nil)
	}
	return utils.Created(c, "satuan berhasil dibuat", s)
}

// UpdateSatuan PUT /api/v1/gudang/satuan/:id
func (h *Controller) UpdateSatuan(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "id satuan tidak valid", nil)
	}
	s, err := h.repo.FindSatuanByID(id)
	if err != nil {
		return utils.Fail(c, fiber.StatusNotFound, "satuan tidak ditemukan", nil)
	}

	var req SatuanRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "payload tidak valid", nil)
	}
	if errs := utils.Validate(req); errs != nil {
		return utils.Fail(c, fiber.StatusUnprocessableEntity, "validasi gagal", errs)
	}
	if req.Nama != s.Nama {
		if existing, err := h.repo.FindSatuanByNama(req.Nama); err == nil && existing.ID != s.ID {
			return utils.Fail(c, fiber.StatusConflict, "nama satuan sudah digunakan", nil)
		}
	}

	s.Nama = req.Nama
	s.Singkatan = req.Singkatan
	if err := h.repo.UpdateSatuan(s); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal memperbarui satuan", nil)
	}
	return utils.OK(c, "satuan berhasil diperbarui", s)
}

// DeleteSatuan DELETE /api/v1/gudang/satuan/:id
func (h *Controller) DeleteSatuan(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "id satuan tidak valid", nil)
	}
	if _, err := h.repo.FindSatuanByID(id); err != nil {
		return utils.Fail(c, fiber.StatusNotFound, "satuan tidak ditemukan", nil)
	}
	if err := h.repo.DeleteSatuan(id); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal menghapus satuan", nil)
	}
	return utils.OK(c, "satuan berhasil dihapus", nil)
}

// ---- Gudang ----

// ListGudang GET /api/v1/gudang?page=&limit=&search=
func (h *Controller) ListGudang(c *fiber.Ctx) error {
	p := utils.PaginationFromContext(c)
	list, total, err := h.repo.ListGudang(p)
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal mengambil daftar gudang", nil)
	}
	return utils.OKWithMeta(c, "daftar gudang berhasil diambil", list, utils.BuildPaginationMeta(p, total))
}

// DetailGudang GET /api/v1/gudang/:id — termasuk daftar rak di gudang tsb.
func (h *Controller) DetailGudang(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "id gudang tidak valid", nil)
	}
	g, err := h.repo.FindGudangByID(id)
	if err != nil {
		return utils.Fail(c, fiber.StatusNotFound, "gudang tidak ditemukan", nil)
	}
	return utils.OK(c, "detail gudang berhasil diambil", g)
}

// CreateGudang POST /api/v1/gudang
func (h *Controller) CreateGudang(c *fiber.Ctx) error {
	var req GudangRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "payload tidak valid", nil)
	}
	if errs := utils.Validate(req); errs != nil {
		return utils.Fail(c, fiber.StatusUnprocessableEntity, "validasi gagal", errs)
	}

	g := &model.Gudang{Nama: req.Nama, Alamat: req.Alamat}
	if err := h.repo.CreateGudang(g); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal membuat gudang", nil)
	}
	return utils.Created(c, "gudang berhasil dibuat", g)
}

// UpdateGudang PUT /api/v1/gudang/:id
func (h *Controller) UpdateGudang(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "id gudang tidak valid", nil)
	}
	g, err := h.repo.FindGudangByID(id)
	if err != nil {
		return utils.Fail(c, fiber.StatusNotFound, "gudang tidak ditemukan", nil)
	}

	var req GudangRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "payload tidak valid", nil)
	}
	if errs := utils.Validate(req); errs != nil {
		return utils.Fail(c, fiber.StatusUnprocessableEntity, "validasi gagal", errs)
	}

	g.Nama = req.Nama
	g.Alamat = req.Alamat
	if err := h.repo.UpdateGudang(g); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal memperbarui gudang", nil)
	}
	return utils.OK(c, "gudang berhasil diperbarui", g)
}

func (h *Controller) DeleteGudang(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "id gudang tidak valid", nil)
	}
	if _, err := h.repo.FindGudangByID(id); err != nil {
		return utils.Fail(c, fiber.StatusNotFound, "gudang tidak ditemukan", nil)
	}

	rakCount, err := h.repo.CountRakByGudang(id)
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal memeriksa rak terkait", nil)
	}
	if rakCount > 0 {
		return utils.Fail(c, fiber.StatusConflict, "gudang masih memiliki rak terdaftar, pindahkan atau hapus rak tersebut terlebih dahulu", nil)
	}

	if err := h.repo.DeleteGudang(id); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal menghapus gudang", nil)
	}
	return utils.OK(c, "gudang berhasil dihapus", nil)
}

// ---- Rak ----

func (h *Controller) ListRak(c *fiber.Ctx) error {
	p := utils.PaginationFromContext(c)
	gudangID, err := strconv.ParseUint(c.Query("gudang_id", "0"), 10, 64)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "gudang_id tidak valid", nil)
	}

	list, total, err := h.repo.ListRak(p, uint(gudangID))
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal mengambil daftar rak", nil)
	}
	return utils.OKWithMeta(c, "daftar rak berhasil diambil", list, utils.BuildPaginationMeta(p, total))
}

// DetailRak GET /api/v1/gudang/rak/:id
func (h *Controller) DetailRak(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "id rak tidak valid", nil)
	}
	rak, err := h.repo.FindRakByID(id)
	if err != nil {
		return utils.Fail(c, fiber.StatusNotFound, "rak tidak ditemukan", nil)
	}
	return utils.OK(c, "detail rak berhasil diambil", rak)
}

// CreateRak POST /api/v1/gudang/rak
func (h *Controller) CreateRak(c *fiber.Ctx) error {
	var req RakRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "payload tidak valid", nil)
	}
	if errs := utils.Validate(req); errs != nil {
		return utils.Fail(c, fiber.StatusUnprocessableEntity, "validasi gagal", errs)
	}

	if _, err := h.repo.FindRakByKode(req.KodeRak); err == nil {
		return utils.Fail(c, fiber.StatusConflict, "kode rak sudah digunakan", nil)
	}
	if _, err := h.repo.FindGudangByID(req.GudangID); err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "gudang tidak ditemukan", nil)
	}

	rak := &model.Rak{KodeRak: req.KodeRak, GudangID: req.GudangID, Kapasitas: req.Kapasitas, Status: "kosong"}
	if err := h.repo.CreateRak(rak); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal membuat rak", nil)
	}
	return utils.Created(c, "rak berhasil dibuat", rak)
}

// UpdateRak PUT /api/v1/gudang/rak/:id
func (h *Controller) UpdateRak(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "id rak tidak valid", nil)
	}
	rak, err := h.repo.FindRakByID(id)
	if err != nil {
		return utils.Fail(c, fiber.StatusNotFound, "rak tidak ditemukan", nil)
	}

	var req UpdateRakRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "payload tidak valid", nil)
	}
	if errs := utils.Validate(req); errs != nil {
		return utils.Fail(c, fiber.StatusUnprocessableEntity, "validasi gagal", errs)
	}

	if req.Kapasitas != nil {
		if *req.Kapasitas < rak.Terisi {
			return utils.Fail(c, fiber.StatusUnprocessableEntity,
				"kapasitas baru tidak boleh lebih kecil dari jumlah unit yang sudah terisi ("+strconv.Itoa(rak.Terisi)+")", nil)
		}
		rak.Kapasitas = *req.Kapasitas
		rak.RecalculateStatus()
	}
	if err := h.repo.UpdateRak(rak); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal memperbarui rak", nil)
	}
	return utils.OK(c, "rak berhasil diperbarui", rak)
}

func (h *Controller) DeleteRak(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "id rak tidak valid", nil)
	}
	rak, err := h.repo.FindRakByID(id)
	if err != nil {
		return utils.Fail(c, fiber.StatusNotFound, "rak tidak ditemukan", nil)
	}
	if rak.Terisi > 0 {
		return utils.Fail(c, fiber.StatusConflict, "rak masih menyimpan unit barang, kosongkan rak terlebih dahulu", nil)
	}

	if err := h.repo.DeleteRak(id); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal menghapus rak", nil)
	}
	return utils.OK(c, "rak berhasil dihapus", nil)
}

func (h *Controller) AdjustRak(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "id rak tidak valid", nil)
	}

	var req AdjustRakRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "payload tidak valid", nil)
	}
	if errs := utils.Validate(req); errs != nil {
		return utils.Fail(c, fiber.StatusUnprocessableEntity, "validasi gagal", errs)
	}

	rak, err := h.repo.AdjustRakTerisi(id, req.Delta)
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal memperbarui kapasitas terisi rak", nil)
	}
	return utils.OK(c, "kapasitas terisi rak berhasil diperbarui", rak)
}

func (h *Controller) Summary(c *fiber.Ctx) error {
	totalGudang, err := h.repo.CountGudang()
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal mengambil ringkasan", nil)
	}
	totalRak, err := h.repo.CountRakAll()
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal mengambil ringkasan", nil)
	}
	penuh, err := h.repo.CountRakByStatus("penuh")
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal mengambil ringkasan", nil)
	}
	kosong, err := h.repo.CountRakByStatus("kosong")
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal mengambil ringkasan", nil)
	}

	return utils.OK(c, "ringkasan rak berhasil diambil", RakSummaryResponse{
		TotalGudang: totalGudang, TotalRak: totalRak, RakTerisiPenuh: penuh, RakKosong: kosong,
	})
}

func (h *Controller) RegisterRoutes(router fiber.Router) {
	g := router.Group("/gudang", middleware.JWTAuth(h.jwtSvc))

	view := middleware.RequirePermission(h.roleRepo, Module, constant.ActionView)
	tambah := middleware.RequirePermission(h.roleRepo, Module, constant.ActionTambah)
	edit := middleware.RequirePermission(h.roleRepo, Module, constant.ActionEdit)

	g.Get("/kategori", view, h.ListKategori)
	g.Post("/kategori", tambah, h.CreateKategori)
	g.Put("/kategori/:id", edit, h.UpdateKategori)
	g.Delete("/kategori/:id", edit, h.DeleteKategori)

	g.Get("/satuan", view, h.ListSatuan)
	g.Post("/satuan", tambah, h.CreateSatuan)
	g.Put("/satuan/:id", edit, h.UpdateSatuan)
	g.Delete("/satuan/:id", edit, h.DeleteSatuan)

	g.Get("/", view, h.ListGudang)
	g.Get("/:id", view, h.DetailGudang)
	g.Post("/", tambah, h.CreateGudang)
	g.Put("/:id", edit, h.UpdateGudang)
	g.Delete("/:id", edit, h.DeleteGudang)

	g.Get("/rak/summary", view, h.Summary)
	g.Get("/rak", view, h.ListRak)
	g.Get("/rak/:id", view, h.DetailRak)
	g.Post("/rak", tambah, h.CreateRak)
	g.Put("/rak/:id", edit, h.UpdateRak)
	g.Delete("/rak/:id", edit, h.DeleteRak)
	g.Patch("/rak/:id/adjust", edit, h.AdjustRak)
}
