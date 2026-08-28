package gudang

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/projsonal/gowms/internal/middleware"
	"github.com/projsonal/gowms/internal/model"
	"github.com/projsonal/gowms/pkg/constant"
	"github.com/projsonal/gowms/pkg/utils"
)

const Module = constant.ModuleManajemenGudang

func normalizeTipeGudang(tipe string) string {
	if tipe == constant.TipeGudangPusat {
		return constant.TipeGudangPusat
	}
	return constant.TipeGudangCabang
}

func parseIDParam(c *fiber.Ctx) (uint, error) {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

func (h *Controller) ListKategori(c *fiber.Ctx) error {
	p := utils.PaginationFromContext(c)
	list, total, err := h.repo.ListKategori(p)
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal mengambil daftar kategori", nil)
	}
	return utils.OKWithMeta(c, "daftar kategori berhasil diambil", list, utils.BuildPaginationMeta(p, total))
}

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

func (h *Controller) ListSatuan(c *fiber.Ctx) error {
	p := utils.PaginationFromContext(c)
	list, total, err := h.repo.ListSatuan(p)
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal mengambil daftar satuan", nil)
	}
	return utils.OKWithMeta(c, "daftar satuan berhasil diambil", list, utils.BuildPaginationMeta(p, total))
}

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

func maskProtectedOne(role string, g *model.Gudang) {
	if role == constant.RoleSuperAdmin || role == constant.RoleAdmin || !g.IsProtected {
		return
	}
	g.Alamat = "*** dilindungi ***"
}

func maskProtected(role string, list []model.Gudang) {
	for i := range list {
		maskProtectedOne(role, &list[i])
	}
}

func (h *Controller) ListGudang(c *fiber.Ctx) error {
	p := utils.PaginationFromContext(c)
	list, total, err := h.repo.ListGudang(p)
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal mengambil daftar gudang", nil)
	}
	roleName, _ := c.Locals(constant.CtxRoleName).(string)
	maskProtected(roleName, list)
	return utils.OKWithMeta(c, "daftar gudang berhasil diambil", list, utils.BuildPaginationMeta(p, total))
}

func (h *Controller) DetailGudang(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "id gudang tidak valid", nil)
	}
	g, err := h.repo.FindGudangByID(id)
	if err != nil {
		return utils.Fail(c, fiber.StatusNotFound, "gudang tidak ditemukan", nil)
	}
	roleName, _ := c.Locals(constant.CtxRoleName).(string)
	maskProtectedOne(roleName, g)
	return utils.OK(c, "detail gudang berhasil diambil", g)
}

func (h *Controller) CreateGudang(c *fiber.Ctx) error {
	var req GudangRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "payload tidak valid", nil)
	}
	if errs := utils.Validate(req); errs != nil {
		return utils.Fail(c, fiber.StatusUnprocessableEntity, "validasi gagal", errs)
	}

	kode := strings.ToUpper(strings.TrimSpace(req.Kode))
	if existing, err := h.repo.FindGudangByKode(kode); err == nil && existing != nil {
		return utils.Fail(c, fiber.StatusConflict, "kode gudang sudah digunakan", nil)
	}

	g := &model.Gudang{
		Nama: req.Nama, Kode: kode, Tipe: normalizeTipeGudang(req.Tipe), Alamat: req.Alamat, PIC: req.PIC, Telepon: req.Telepon, Kapasitas: req.Kapasitas,
		Latitude: req.Latitude, Longitude: req.Longitude,
	}
	if err := h.repo.CreateGudang(g); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal membuat gudang", nil)
	}
	return utils.Created(c, "gudang berhasil dibuat", g)
}

func (h *Controller) UpdateGudang(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "id gudang tidak valid", nil)
	}
	g, err := h.repo.FindGudangByID(id)
	if err != nil {
		return utils.Fail(c, fiber.StatusNotFound, "gudang tidak ditemukan", nil)
	}
	if g.IsProtected {
		return utils.Fail(c, fiber.StatusForbidden,
			"data ini dikunci (Protect) oleh super admin — buka kuncinya dulu sebelum diubah", nil)
	}

	var req GudangRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "payload tidak valid", nil)
	}
	if errs := utils.Validate(req); errs != nil {
		return utils.Fail(c, fiber.StatusUnprocessableEntity, "validasi gagal", errs)
	}

	kode := strings.ToUpper(strings.TrimSpace(req.Kode))
	if kode != g.Kode {
		if existing, err := h.repo.FindGudangByKode(kode); err == nil && existing != nil && existing.ID != g.ID {
			return utils.Fail(c, fiber.StatusConflict, "kode gudang sudah digunakan", nil)
		}
	}

	g.Nama = req.Nama
	g.Kode = kode
	g.Tipe = normalizeTipeGudang(req.Tipe)
	g.Alamat = req.Alamat
	g.PIC = req.PIC
	g.Telepon = req.Telepon
	g.Kapasitas = req.Kapasitas
	g.Latitude = req.Latitude
	g.Longitude = req.Longitude
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
	g, err := h.repo.FindGudangByID(id)
	if err != nil {
		return utils.Fail(c, fiber.StatusNotFound, "gudang tidak ditemukan", nil)
	}
	if g.IsProtected {
		return utils.Fail(c, fiber.StatusForbidden,
			"data ini dikunci (Protect) oleh super admin — buka kuncinya dulu sebelum dihapus", nil)
	}

	if err := h.repo.DeleteGudang(id); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal menghapus gudang", nil)
	}
	return utils.OK(c, "gudang berhasil dihapus", nil)
}

func (h *Controller) ProtectGudang(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "id gudang tidak valid", nil)
	}
	var req ProtectRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "payload tidak valid", nil)
	}
	if errs := utils.Validate(req); errs != nil {
		return utils.Fail(c, fiber.StatusUnprocessableEntity, "validasi gagal", errs)
	}
	g, err := h.repo.FindGudangByID(id)
	if err != nil {
		return utils.Fail(c, fiber.StatusNotFound, "gudang tidak ditemukan", nil)
	}
	g.IsProtected = *req.IsProtected
	if err := h.repo.UpdateGudang(g); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal mengubah status proteksi", nil)
	}
	return utils.OK(c, "status proteksi berhasil diubah", g)
}

func (h *Controller) RegisterRoutes(router fiber.Router) {
	g := router.Group("/gudang", middleware.JWTAuth(h.jwtSvc))

	view := middleware.RequirePermission(h.roleRepo, Module, constant.ActionView)
	tambah := middleware.RequirePermission(h.roleRepo, Module, constant.ActionTambah)
	edit := middleware.RequirePermission(h.roleRepo, Module, constant.ActionEdit)
	onlySuperAdmin := middleware.RequireRole(constant.RoleSuperAdmin)
	onlyStaff := middleware.RequireRole(constant.RoleSuperAdmin, constant.RoleAdmin)

	g.Get("/kategori", view, h.ListKategori)
	g.Post("/kategori", tambah, h.CreateKategori)
	g.Put("/kategori/:id", edit, h.UpdateKategori)
	g.Delete("/kategori/:id", onlyStaff, edit, h.DeleteKategori)

	g.Get("/satuan", view, h.ListSatuan)
	g.Post("/satuan", tambah, h.CreateSatuan)
	g.Put("/satuan/:id", edit, h.UpdateSatuan)
	g.Delete("/satuan/:id", onlyStaff, edit, h.DeleteSatuan)

	g.Get("/", view, h.ListGudang)
	g.Get("/:id", view, h.DetailGudang)
	g.Post("/", tambah, h.CreateGudang)
	g.Put("/:id", edit, h.UpdateGudang)
	g.Delete("/:id", onlyStaff, edit, h.DeleteGudang)
	g.Patch("/:id/protect", onlySuperAdmin, h.ProtectGudang)
}
