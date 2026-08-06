package stock_opname

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/projsonal/gostock/internal/middleware"
	"github.com/projsonal/gostock/internal/model"
	soRepo "github.com/projsonal/gostock/internal/repositories/stockOpname"
	"github.com/projsonal/gostock/pkg/constant"
	"github.com/projsonal/gostock/pkg/utils"
)

const Module = constant.ModuleStockOpname

func parseIDParam(c *fiber.Ctx) (uint, error) {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

func generateNomorSO() string {
	return fmt.Sprintf("SO-%d-%d", time.Now().Year(), time.Now().UnixNano()%100000)
}

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

func toItemInputs(items []ItemRequest) []soRepo.ItemInput {
	out := make([]soRepo.ItemInput, 0, len(items))
	for _, it := range items {
		out = append(out, soRepo.ItemInput{
			BarangID: it.BarangID, RakID: it.RakID, StokFisik: it.StokFisik, Catatan: it.Catatan,
		})
	}
	return out
}

// List GET /stock-opname?page=&limit=&search=&status=&gudang_id=
func (h *Controller) List(c *fiber.Ctx) error {
	p := utils.PaginationFromContext(c)
	gudangID, _ := strconv.ParseUint(c.Query("gudang_id", "0"), 10, 64)
	f := soRepo.Filter{Status: c.Query("status", ""), GudangID: uint(gudangID)}

	list, total, err := h.repo.List(p, f)
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal mengambil daftar stock opname", nil)
	}
	return utils.OKWithMeta(c, "daftar stock opname berhasil diambil", list, utils.BuildPaginationMeta(p, total))
}

// Detail GET /stock-opname/:id
func (h *Controller) Detail(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "id stock opname tidak valid", nil)
	}
	so, err := h.repo.FindByID(id)
	if err != nil {
		return utils.Fail(c, fiber.StatusNotFound, constant.ErrSOTidakDitemukan, nil)
	}
	return utils.OK(c, "detail stock opname berhasil diambil", so)
}

// Create POST /stock-opname — StokSistem tiap item di-snapshot otomatis
// dari Barang.Stok saat ini oleh repository (bukan dari input client),
// operator hanya mengisi StokFisik hasil hitung manual.
func (h *Controller) Create(c *fiber.Ctx) error {
	var req SORequest
	if !utils.ParseAndValidate(c, &req) {
		return nil
	}
	if _, err := h.gudangRepo.FindGudangByID(req.GudangID); err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "gudang tidak ditemukan", nil)
	}
	if err := h.validateItems(req.Items); err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	so := &model.StockOpname{
		NomorOpname: generateNomorSO(),
		GudangID:    req.GudangID,
		Status:      constant.StatusSODraft,
		Tanggal:     req.Tanggal,
		Catatan:     req.Catatan,
	}
	if err := h.repo.Create(so, toItemInputs(req.Items)); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal membuat dokumen stock opname", nil)
	}
	return utils.Created(c, "dokumen stock opname berhasil dibuat", so)
}

func (h *Controller) requireDraft(id uint) (*model.StockOpname, error) {
	so, err := h.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if so.Status != constant.StatusSODraft {
		return nil, errors.New(constant.ErrSOBukanDraft)
	}
	return so, nil
}

// Update PUT /stock-opname/:id — hanya boleh selama status masih draft.
func (h *Controller) Update(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "id stock opname tidak valid", nil)
	}
	so, err := h.requireDraft(id)
	if err != nil {
		return utils.Fail(c, fiber.StatusConflict, err.Error(), nil)
	}

	var req SORequest
	if !utils.ParseAndValidate(c, &req) {
		return nil
	}
	if _, err := h.gudangRepo.FindGudangByID(req.GudangID); err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "gudang tidak ditemukan", nil)
	}
	if err := h.validateItems(req.Items); err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	so.GudangID = req.GudangID
	so.Tanggal = req.Tanggal
	so.Catatan = req.Catatan
	if err := h.repo.Update(so, toItemInputs(req.Items)); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal memperbarui dokumen stock opname", nil)
	}
	return utils.OK(c, "dokumen stock opname berhasil diperbarui", so)
}

// Delete DELETE /stock-opname/:id — hanya boleh selama status draft.
func (h *Controller) Delete(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "id stock opname tidak valid", nil)
	}
	if _, err := h.requireDraft(id); err != nil {
		return utils.Fail(c, fiber.StatusConflict, err.Error(), nil)
	}
	if err := h.repo.Delete(id); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal menghapus dokumen stock opname", nil)
	}
	return utils.OK(c, "dokumen stock opname berhasil dihapus", nil)
}

// Complete PATCH /stock-opname/:id/selesai — menerapkan selisih hasil
// hitung fisik ke stok barang (dan isi rak bila item terkait rak).
func (h *Controller) Complete(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "id stock opname tidak valid", nil)
	}
	userID, _ := c.Locals(constant.CtxUserID).(uint)

	so, err := h.repo.Complete(id, userID)
	if err != nil {
		return utils.Fail(c, fiber.StatusConflict, err.Error(), nil)
	}
	return utils.OK(c, "stock opname berhasil diselesaikan, selisih telah diterapkan ke stok", so)
}

// Batalkan PATCH /stock-opname/:id/batalkan
func (h *Controller) Batalkan(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "id stock opname tidak valid", nil)
	}
	so, err := h.repo.Batalkan(id)
	if err != nil {
		return utils.Fail(c, fiber.StatusConflict, err.Error(), nil)
	}
	return utils.OK(c, "dokumen stock opname berhasil dibatalkan", so)
}

// Summary GET /stock-opname/summary
func (h *Controller) Summary(c *fiber.Ctx) error {
	total, err := h.repo.CountByStatus("")
	draft, err2 := h.repo.CountByStatus(constant.StatusSODraft)
	selesai, err3 := h.repo.CountByStatus(constant.StatusSOSelesai)
	if err != nil || err2 != nil || err3 != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal mengambil ringkasan", nil)
	}
	return utils.OK(c, "ringkasan stock opname berhasil diambil", SummaryResponse{
		TotalDokumen: total, Draft: draft, Selesai: selesai,
	})
}

// RegisterRoutes mendaftarkan endpoint modul "Stock Opname".
func (h *Controller) RegisterRoutes(router fiber.Router) {
	g := router.Group("/stock-opname", middleware.JWTAuth(h.jwtSvc))

	view := middleware.RequirePermission(h.roleRepo, Module, constant.ActionView)
	tambah := middleware.RequirePermission(h.roleRepo, Module, constant.ActionTambah)
	edit := middleware.RequirePermission(h.roleRepo, Module, constant.ActionEdit)

	g.Get("/summary", view, h.Summary)
	g.Get("/", view, h.List)
	g.Get("/:id", view, h.Detail)
	g.Post("/", tambah, h.Create)
	g.Put("/:id", edit, h.Update)
	g.Delete("/:id", edit, h.Delete)
	g.Patch("/:id/selesai", edit, h.Complete)
	g.Patch("/:id/batalkan", edit, h.Batalkan)
}
