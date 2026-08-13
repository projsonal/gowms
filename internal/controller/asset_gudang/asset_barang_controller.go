package assetgudang

import (
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"github.com/projsonal/gowms/internal/middleware"
	"github.com/projsonal/gowms/internal/model"
	assetRepo "github.com/projsonal/gowms/internal/repositories/asset"
	"github.com/projsonal/gowms/pkg/constant"
	"github.com/projsonal/gowms/pkg/utils"
)

func parseIDParam(c *fiber.Ctx) (uint, error) {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

// List GET /aset?jenis_aset=&gudang_id=&status=&page=&limit=&search=
func (h *Controller) List(c *fiber.Ctx) error {
	p := utils.PaginationFromContext(c)
	gudangID, _ := strconv.ParseUint(c.Query("gudang_id", "0"), 10, 64)
	f := assetRepo.Filter{
		JenisAset: c.Query("jenis_aset", ""),
		GudangID:  uint(gudangID),
		Status:    c.Query("status", ""),
	}
	list, total, err := h.repo.List(p, f)
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal mengambil daftar aset", nil)
	}
	return utils.OKWithMeta(c, "daftar aset berhasil diambil", list, utils.BuildPaginationMeta(p, total))
}

// Detail GET /aset/:id
func (h *Controller) Detail(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "Data id aset tidak valid", nil)
	}
	a, err := h.repo.FindByID(id)
	if err != nil {
		return utils.Fail(c, fiber.StatusNotFound, "Data aset gudang tidak ditemukan", nil)
	}
	return utils.OK(c, "detail aset berhasil diambil", a)
}

// Summary GET /aset/summary — ringkasan jumlah aset per jenis, dipakai
// kartu ringkasan di halaman Manajemen Aset (mirip Task Summary lama).
func (h *Controller) Summary(c *fiber.Ctx) error {
	tiang, _ := h.repo.CountByJenis(constant.JenisAsetTiang)
	odc, _ := h.repo.CountByJenis(constant.JenisAsetODC)
	ont, _ := h.repo.CountByJenis(constant.JenisAsetONT)
	odp, _ := h.repo.CountByJenis(constant.JenisAsetODP)
	olt, _ := h.repo.CountByJenis(constant.JenisAsetOLT)
	transportasi, _ := h.repo.CountByJenis(constant.JenisAsetTransportasi)
	return utils.OK(c, "ringkasan aset berhasil diambil", SummaryResponse{
		Tiang: tiang, Odc: odc, Ont: ont, Odp: odp, Olt: olt, Transportasi: transportasi,
		Total: tiang + odc + ont + odp + olt + transportasi,
	})
}

// Create POST /aset — label RSD / kode BA dibuat OTOMATIS oleh server,
// TIDAK boleh dikirim dari klien, supaya format & urutannya konsisten:
//   - tiang/odc/ont/odp/olt: "{KodeGudang}-RSD-{nomor urut per gudang}",
//     wajib menyertakan koordinat (latitude & longitude).
//   - transportasi: "BA-{nomor urut global}", tanpa koordinat.
func (h *Controller) Create(c *fiber.Ctx) error {
	var req AssetRequest
	if !utils.ParseAndValidate(c, &req) {
		return nil
	}

	gudang, err := h.gudangRepo.FindGudangByID(req.GudangID)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "gudang tidak ditemukan", nil)
	}

	a := &model.Asset{
		Nama:       req.Nama,
		JenisAset:  req.JenisAset,
		GudangID:   req.GudangID,
		Keterangan: req.Keterangan,
		Status:     "aktif",
	}

	if model.JenisAsetPunyaKoordinat(req.JenisAset) {
		if req.Latitude == nil || req.Longitude == nil {
			return utils.Fail(c, fiber.StatusUnprocessableEntity,
				"latitude dan longitude wajib diisi untuk jenis aset ini", nil)
		}
		if gudang.Kode == "" {
			return utils.Fail(c, fiber.StatusUnprocessableEntity,
				"gudang belum punya kode — isi kode gudang dulu sebelum menambah aset", nil)
		}
		nomor, err := h.repo.NextRSDNumber(req.GudangID)
		if err != nil {
			return utils.Fail(c, fiber.StatusInternalServerError, "gagal membuat label RSD", nil)
		}
		a.LabelRSD = fmt.Sprintf("%s-RSD-%04d", gudang.Kode, nomor)
		a.Latitude = req.Latitude
		a.Longitude = req.Longitude
	} else {
		nomor, err := h.repo.NextBANumber()
		if err != nil {
			return utils.Fail(c, fiber.StatusInternalServerError, "gagal membuat kode BA", nil)
		}
		a.KodeBA = fmt.Sprintf("BA-%04d", nomor)
	}

	if err := h.repo.Create(a); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal membuat aset", nil)
	}
	created, _ := h.repo.FindByID(a.ID)
	return utils.Created(c, "aset berhasil dibuat", created)
}

// Update PUT /aset/:id — label RSD / kode BA TIDAK BISA diubah lewat
// endpoint ini (identitas aset harus tetap), hanya data deskriptif &
// (untuk aset berkoordinat) titik lokasinya.
func (h *Controller) Update(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "Data id aset tidak valid", nil)
	}
	a, err := h.repo.FindByID(id)
	if err != nil {
		return utils.Fail(c, fiber.StatusNotFound, "Data aset gudang tidak ditemukan", nil)
	}

	var req AssetRequest
	if !utils.ParseAndValidate(c, &req) {
		return nil
	}

	a.Nama = req.Nama
	a.Keterangan = req.Keterangan
	if model.JenisAsetPunyaKoordinat(a.JenisAset) {
		if req.Latitude == nil || req.Longitude == nil {
			return utils.Fail(c, fiber.StatusUnprocessableEntity,
				"latitude dan longitude wajib diisi untuk jenis aset ini", nil)
		}
		a.Latitude = req.Latitude
		a.Longitude = req.Longitude
	}
	if err := h.repo.Update(a); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal memperbarui aset", nil)
	}
	return utils.OK(c, "aset berhasil diperbarui", a)
}

// UpdateStatus PATCH /aset/:id/status
func (h *Controller) UpdateStatus(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "id aset tidak valid", nil)
	}
	a, err := h.repo.FindByID(id)
	if err != nil {
		return utils.Fail(c, fiber.StatusNotFound, "aset tidak ditemukan", nil)
	}

	var req UpdateStatusRequest
	if !utils.ParseAndValidate(c, &req) {
		return nil
	}
	a.Status = req.Status
	if err := h.repo.Update(a); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal memperbarui status aset", nil)
	}
	return utils.OK(c, "status aset berhasil diperbarui", a)
}

// Delete DELETE /aset/:id
func (h *Controller) Delete(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "id aset tidak valid", nil)
	}
	if _, err := h.repo.FindByID(id); err != nil {
		return utils.Fail(c, fiber.StatusNotFound, "aset tidak ditemukan", nil)
	}
	if err := h.repo.Delete(id); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal menghapus aset", nil)
	}
	return utils.OK(c, "aset berhasil dihapus", nil)
}

func (h *Controller) RegisterRoutes(router fiber.Router) {
	g := router.Group("/aset", middleware.JWTAuth(h.jwtSvc))

	view := middleware.RequirePermission(h.roleRepo, Module, constant.ActionView)
	tambah := middleware.RequirePermission(h.roleRepo, Module, constant.ActionTambah)
	edit := middleware.RequirePermission(h.roleRepo, Module, constant.ActionEdit)
	onlyStaff := middleware.RequireRole(constant.RoleSuperAdmin, constant.RoleAdmin)

	g.Get("/summary", view, h.Summary)
	g.Get("/", view, h.List)
	g.Get("/:id", view, h.Detail)
	g.Post("/", tambah, onlyStaff, h.Create)
	g.Put("/:id", edit, onlyStaff, h.Update)
	g.Patch("/:id/status", edit, h.UpdateStatus)
	g.Delete("/:id", onlyStaff, edit, h.Delete)
}
