package barang_rusak

import (
	"io"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	notification "github.com/projsonal/gowms/internal/controller/notifikasi"
	"github.com/projsonal/gowms/internal/middleware"
	"github.com/projsonal/gowms/internal/model"
	"github.com/projsonal/gowms/pkg/constant"
	"github.com/projsonal/gowms/pkg/utils"

	barangRusakRepo "github.com/projsonal/gowms/internal/repositories/barang_rusak"
)

func parseIDParam(c *fiber.Ctx) (uint, error) {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

func (h *Controller) List(c *fiber.Ctx) error {
	p := utils.PaginationFromContext(c)
	f := barangRusakRepo.Filter{Status: c.Query("status", "")}
	list, total, err := h.repo.List(p, f)
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal mengambil daftar barang rusak", nil)
	}
	return utils.OKWithMeta(c, "daftar barang rusak berhasil diambil", list, utils.BuildPaginationMeta(p, total))
}

func (h *Controller) Detail(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "Id barang rusak tidak valid", nil)
	}
	b, err := h.repo.FindByID(id)
	if err != nil {
		return utils.Fail(c, fiber.StatusNotFound, "Data barang rusak tidak ditemukan", nil)
	}
	return utils.OK(c, "detail barang rusak berhasil diambil", b)
}

func (h *Controller) Summary(c *fiber.Ctx) error {
	pengecekan, _ := h.repo.CountByStatus(constant.StatusPengecekan)
	retur, _ := h.repo.CountByStatus(constant.StatusRetur)
	rusak, _ := h.repo.CountByStatus(constant.StatusRusak)
	return utils.OK(c, "ringkasan barang rusak berhasil diambil", SummaryResponse{
		Pengecekan: pengecekan, Retur: retur, Rusak: rusak,
		Total: pengecekan + retur + rusak,
	})
}

func (h *Controller) Create(c *fiber.Ctx) error {
	var req BarangRusakRequest
	if !utils.ParseAndValidate(c, &req) {
		return nil
	}
	if req.BarangID != nil {
		if _, err := h.barangRepo.FindByID(*req.BarangID); err != nil {
			return utils.Fail(c, fiber.StatusBadRequest, "barang tidak ditemukan", nil)
		}
	}

	userID, _ := c.Locals(constant.CtxUserID).(uint)
	b := &model.BarangRusak{
		BarangID:       req.BarangID,
		LabelBarang:    req.LabelBarang,
		NamaBarang:     req.NamaBarang,
		SerialNumber:   req.SerialNumber,
		Keterangan:     req.Keterangan,
		Status:         constant.StatusPengecekan,
		DilaporkanOleh: userID,
	}
	if err := h.repo.Create(b); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal mencatat laporan barang rusak", nil)
	}
	notification.Notify(h.notifRepo, "barang_rusak",
		"Laporan Barang Rusak Baru",
		b.NamaBarang+" ("+b.LabelBarang+") dilaporkan rusak, menunggu pengecekan fisik.",
		"/home/barang-rusak", nil, "admin")
	notification.Notify(h.notifRepo, "barang_rusak",
		"Laporan Barang Rusak Baru",
		b.NamaBarang+" ("+b.LabelBarang+") dilaporkan rusak, menunggu pengecekan fisik.",
		"/home/barang-rusak", nil, constant.RoleSuperAdmin)
	created, _ := h.repo.FindByID(b.ID)
	return utils.Created(c, "laporan barang rusak berhasil dibuat, menunggu pengecekan fisik", created)
}

func (h *Controller) Update(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "id tidak valid", nil)
	}
	b, err := h.repo.FindByID(id)
	if err != nil {
		return utils.Fail(c, fiber.StatusNotFound, "data barang rusak tidak ditemukan", nil)
	}
	if b.Status != constant.StatusPengecekan {
		return utils.Fail(c, fiber.StatusForbidden, "data yang sudah diperiksa tidak bisa diubah", nil)
	}

	var req BarangRusakRequest
	if !utils.ParseAndValidate(c, &req) {
		return nil
	}
	if req.BarangID != nil {
		if _, err := h.barangRepo.FindByID(*req.BarangID); err != nil {
			return utils.Fail(c, fiber.StatusBadRequest, "barang tidak ditemukan", nil)
		}
	}
	b.BarangID = req.BarangID
	b.LabelBarang = req.LabelBarang
	b.NamaBarang = req.NamaBarang
	b.SerialNumber = req.SerialNumber
	b.Keterangan = req.Keterangan
	if err := h.repo.Update(b); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal memperbarui laporan barang rusak", nil)
	}
	return utils.OK(c, "laporan barang rusak berhasil diperbarui", b)
}

func (h *Controller) Inspeksi(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "id tidak valid", nil)
	}
	b, err := h.repo.FindByID(id)
	if err != nil {
		return utils.Fail(c, fiber.StatusNotFound, "data barang rusak tidak ditemukan", nil)
	}
	if b.Status != constant.StatusPengecekan {
		return utils.Fail(c, fiber.StatusConflict, "data ini sudah pernah diperiksa", nil)
	}

	var req InspeksiRequest
	if !utils.ParseAndValidate(c, &req) {
		return nil
	}

	userID, _ := c.Locals(constant.CtxUserID).(uint)
	now := time.Now()
	b.JenisBarang = req.JenisBarang
	b.Status = req.JenisBarang
	b.DicekOleh = &userID
	b.DicekPada = &now
	if err := h.repo.Update(b); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal menyimpan hasil pengecekan", nil)
	}
	pelaporID := b.DilaporkanOleh
	hasilLabel := "bisa diretur ke supplier"
	if req.JenisBarang == constant.StatusRusak {
		hasilLabel = "rusak total (tidak bisa diretur)"
	}
	notification.Notify(h.notifRepo, "barang_rusak",
		"Hasil Pengecekan Barang Rusak",
		b.NamaBarang+" ("+b.LabelBarang+") sudah diperiksa: "+hasilLabel+".",
		"/home/barang-rusak", &pelaporID, "")
	return utils.OK(c, "hasil pengecekan berhasil disimpan", b)
}

func (h *Controller) Delete(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "id tidak valid", nil)
	}
	if _, err := h.repo.FindByID(id); err != nil {
		return utils.Fail(c, fiber.StatusNotFound, "data barang rusak tidak ditemukan", nil)
	}
	if err := h.repo.Delete(id); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal menghapus data barang rusak", nil)
	}
	return utils.OK(c, "data barang rusak berhasil dihapus", nil)
}

func (h *Controller) UploadFoto(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "id tidak valid", nil)
	}
	b, err := h.repo.FindByID(id)
	if err != nil {
		return utils.Fail(c, fiber.StatusNotFound, "data barang rusak tidak ditemukan", nil)
	}

	file, err := c.FormFile("foto")
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "file foto tidak ditemukan (field: foto)", nil)
	}
	const maxFotoSize = 2 * 1024 * 1024
	if file.Size > maxFotoSize {
		return utils.Fail(c, fiber.StatusBadRequest, "ukuran file maksimal 2MB", nil)
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	var contentType string
	switch ext {
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".png":
		contentType = "image/png"
	default:
		return utils.Fail(c, fiber.StatusBadRequest, "format file harus jpg, jpeg, atau png", nil)
	}

	opened, err := file.Open()
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal membaca file foto", nil)
	}
	defer opened.Close()
	data, err := io.ReadAll(opened)
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal membaca file foto", nil)
	}

	b.FotoData = data
	b.FotoContentType = contentType
	if err := h.repo.Update(b); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal menyimpan foto bukti", nil)
	}
	return utils.OK(c, "foto bukti berhasil diunggah", b)
}

func (h *Controller) ServeFoto(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "id tidak valid", nil)
	}
	b, err := h.repo.FindByID(id)
	if err != nil {
		return utils.Fail(c, fiber.StatusNotFound, "data barang rusak tidak ditemukan", nil)
	}
	if len(b.FotoData) == 0 {
		return utils.Fail(c, fiber.StatusNotFound, "belum ada foto bukti yang diunggah", nil)
	}
	c.Set("Content-Type", b.FotoContentType)
	c.Set("Cache-Control", "private, max-age=86400")
	return c.Send(b.FotoData)
}

func (h *Controller) RegisterRoutes(router fiber.Router) {
	g := router.Group("/barang-rusak", middleware.JWTAuth(h.jwtSvc))

	view := middleware.RequirePermission(h.roleRepo, Module, constant.ActionView)
	tambah := middleware.RequirePermission(h.roleRepo, Module, constant.ActionTambah)
	edit := middleware.RequirePermission(h.roleRepo, Module, constant.ActionEdit)
	onlyStaff := middleware.RequireRole(constant.RoleSuperAdmin, constant.RoleAdmin)

	g.Get("/summary", view, h.Summary)
	g.Get("/", view, h.List)
	g.Get("/:id", view, h.Detail)
	g.Post("/", tambah, h.Create)
	g.Put("/:id", edit, h.Update)
	g.Post("/:id/foto", edit, h.UploadFoto)
	g.Get("/:id/foto", view, h.ServeFoto)
	g.Patch("/:id/inspeksi", edit, onlyStaff, h.Inspeksi)
	g.Delete("/:id", onlyStaff, edit, h.Delete)
}
