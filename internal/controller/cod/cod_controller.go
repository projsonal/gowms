package cod

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/projsonal/gowms/internal/middleware"
	"github.com/projsonal/gowms/internal/model"
	codRepo "github.com/projsonal/gowms/internal/repositories/cod"
	"github.com/projsonal/gowms/pkg/constant"
	"github.com/projsonal/gowms/pkg/utils"
)

const Module = constant.ModuleCOD

func parseIDParam(c *fiber.Ctx) (uint, error) {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

func (h *Controller) List(c *fiber.Ctx) error {
	p := utils.PaginationFromContext(c)
	f := codRepo.Filter{Status: c.Query("status", "")}

	list, total, err := h.repo.List(p, f)
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal mengambil daftar transaksi COD", nil)
	}
	return utils.OKWithMeta(c, "daftar transaksi COD berhasil diambil", list, utils.BuildPaginationMeta(p, total))
}

func (h *Controller) Summary(c *fiber.Ctx) error {
	total, err := h.repo.CountByStatus("")
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal mengambil ringkasan COD", nil)
	}
	lunas, _ := h.repo.CountByStatus("lunas")
	menunggu, _ := h.repo.CountByStatus("menunggu")
	bermasalah, _ := h.repo.CountByStatus("bermasalah")
	nominal, _ := h.repo.SumNominal()

	return utils.OK(c, "ringkasan COD berhasil diambil", SummaryResponse{
		Total: total, Lunas: lunas, Menunggu: menunggu, Bermasalah: bermasalah, TotalNominal: nominal,
	})
}

func parseTanggal(raw string) (time.Time, error) {
	return time.Parse("2006-01-02", raw)
}

func (h *Controller) Create(c *fiber.Ctx) error {
	var req CodRequest
	if !utils.ParseAndValidate(c, &req) {
		return nil
	}
	tgl, err := parseTanggal(req.Tanggal)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "format tanggal tidak valid, gunakan YYYY-MM-DD", nil)
	}
	if _, err := h.repo.FindByKode(req.Kode); err == nil {
		return utils.Fail(c, fiber.StatusConflict, "kode COD sudah digunakan", nil)
	}

	row := &model.CodTransaction{
		Kode: req.Kode, Pelanggan: req.Pelanggan, Nominal: req.Nominal,
		Kurir: req.Kurir, Tanggal: tgl, Status: req.Status,
	}
	if err := h.repo.Create(row); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal membuat transaksi COD", nil)
	}
	return utils.Created(c, "transaksi COD berhasil dibuat", row)
}

func (h *Controller) Update(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "id tidak valid", nil)
	}
	row, err := h.repo.FindByID(id)
	if err != nil {
		return utils.Fail(c, fiber.StatusNotFound, "transaksi COD tidak ditemukan", nil)
	}
	if row.IsProtected {
		return utils.Fail(c, fiber.StatusForbidden,
			"data ini dikunci (Protect) oleh super admin — buka kuncinya dulu sebelum diubah", nil)
	}

	var req CodRequest
	if !utils.ParseAndValidate(c, &req) {
		return nil
	}
	tgl, err := parseTanggal(req.Tanggal)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "format tanggal tidak valid, gunakan YYYY-MM-DD", nil)
	}

	row.Kode = req.Kode
	row.Pelanggan = req.Pelanggan
	row.Nominal = req.Nominal
	row.Kurir = req.Kurir
	row.Tanggal = tgl
	row.Status = req.Status
	if err := h.repo.Update(row); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal memperbarui transaksi COD", nil)
	}
	return utils.OK(c, "transaksi COD berhasil diperbarui", row)
}

func (h *Controller) Delete(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "id tidak valid", nil)
	}
	row, err := h.repo.FindByID(id)
	if err != nil {
		return utils.Fail(c, fiber.StatusNotFound, "transaksi COD tidak ditemukan", nil)
	}
	if row.IsProtected {
		return utils.Fail(c, fiber.StatusForbidden,
			"data ini dikunci (Protect) oleh super admin — buka kuncinya dulu sebelum dihapus", nil)
	}
	if err := h.repo.Delete(id); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal menghapus transaksi COD", nil)
	}
	return utils.OK(c, "transaksi COD berhasil dihapus", nil)
}

func (h *Controller) Protect(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "id tidak valid", nil)
	}
	var req ProtectRequest
	if !utils.ParseAndValidate(c, &req) {
		return nil
	}
	row, err := h.repo.FindByID(id)
	if err != nil {
		return utils.Fail(c, fiber.StatusNotFound, "transaksi COD tidak ditemukan", nil)
	}
	row.IsProtected = *req.IsProtected
	if err := h.repo.Update(row); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal mengubah status proteksi", nil)
	}
	return utils.OK(c, "status proteksi berhasil diubah", row)
}

func (h *Controller) RegisterRoutes(router fiber.Router) {
	g := router.Group("/cod", middleware.JWTAuth(h.jwtSvc))

	view := middleware.RequirePermission(h.roleRepo, Module, constant.ActionView)
	tambah := middleware.RequirePermission(h.roleRepo, Module, constant.ActionTambah)
	edit := middleware.RequirePermission(h.roleRepo, Module, constant.ActionEdit)

	onlyStaff := middleware.RequireRole(constant.RoleSuperAdmin, constant.RoleAdmin)
	onlySuperAdmin := middleware.RequireRole(constant.RoleSuperAdmin)

	g.Get("/summary", view, h.Summary)
	g.Get("/", view, h.List)
	g.Post("/", tambah, h.Create)
	g.Put("/:id", edit, h.Update)
	g.Delete("/:id", onlyStaff, edit, h.Delete)
	g.Patch("/:id/protect", onlySuperAdmin, h.Protect)
}
