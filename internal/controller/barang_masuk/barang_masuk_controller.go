package barang_masuk

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/projsonal/gowms/internal/middleware"
	"github.com/projsonal/gowms/internal/model"
	bmRepo "github.com/projsonal/gowms/internal/repositories/barang_masuk"
	"github.com/projsonal/gowms/pkg/constant"
	"github.com/projsonal/gowms/pkg/utils"
)

const Module = constant.ModuleBarangMasuk

const msgIdBM = "id barang masuk tidak valid"

func parseIDParam(c *fiber.Ctx) (uint, error) {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

func generateNomorBM() string {
	return fmt.Sprintf("BM-%d-%d", time.Now().Year(), time.Now().UnixNano()%100000)
}

func (h *Controller) validateItems(req BMRequest) error {
	po, err := h.validatePurchaseOrder(req.PurchaseOrderID)
	if err != nil {
		return err
	}
	return h.validateItemRequests(req.Items, po)
}

func (h *Controller) validatePurchaseOrder(poID *uint) (*model.PurchaseOrder, error) {
	if poID == nil {
		return nil, nil
	}
	po, err := h.poRepo.FindByID(*poID)
	if err != nil {
		return nil, errors.New("purchase order tidak ditemukan")
	}
	if po.Status != constant.StatusPODisetujui {
		return nil, errors.New(constant.ErrPOTidakDisetujui)
	}
	return po, nil
}

func (h *Controller) validateItemRequests(items []ItemRequest, po *model.PurchaseOrder) error {
	for _, it := range items {
		if err := h.validateItem(it, po); err != nil {
			return err
		}
	}
	return nil
}

func (h *Controller) validateItem(it ItemRequest, po *model.PurchaseOrder) error {
	if _, err := h.barangRepo.FindByID(it.BarangID); err != nil {
		return fmt.Errorf("barang id %d tidak ditemukan", it.BarangID)
	}
	if it.RakID != nil {
		if _, err := h.gudangRepo.FindRakByID(*it.RakID); err != nil {
			return fmt.Errorf("rak id %d tidak ditemukan", *it.RakID)
		}
	}
	if po != nil {
		if err := h.validateItemQtyForPO(it, po); err != nil {
			return err
		}
	}
	return nil
}

func (h *Controller) validateItemQtyForPO(it ItemRequest, po *model.PurchaseOrder) error {
	for _, poItem := range po.Items {
		if poItem.BarangID == it.BarangID && it.Qty > poItem.SisaDiterima() {
			return fmt.Errorf("qty barang id %d melebihi sisa yang belum diterima pada PO (sisa: %d)",
				it.BarangID, poItem.SisaDiterima())
		}
	}
	return nil
}

func toItemModels(items []ItemRequest) []model.BarangMasukItem {
	out := make([]model.BarangMasukItem, 0, len(items))
	for _, it := range items {
		out = append(out, model.BarangMasukItem{
			BarangID: it.BarangID, RakID: it.RakID, Qty: it.Qty, HargaSatuan: it.HargaSatuan,
		})
	}
	return out
}

// List GET /barang-masuk?page=&limit=&search=&status=&gudang_id=&purchase_order_id=
func (h *Controller) List(c *fiber.Ctx) error {
	p := utils.PaginationFromContext(c)
	gudangID, _ := strconv.ParseUint(c.Query("gudang_id", "0"), 10, 64)
	poID, _ := strconv.ParseUint(c.Query("purchase_order_id", "0"), 10, 64)
	kategoriID, _ := strconv.ParseUint(c.Query("kategori_id", "0"), 10, 64)
	f := bmRepo.Filter{
		Status:          c.Query("status", ""),
		GudangID:        uint(gudangID),
		PurchaseOrderID: uint(poID),
		KategoriID:      uint(kategoriID),
	}

	list, total, err := h.repo.List(p, f)
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal mengambil daftar barang masuk", nil)
	}
	return utils.OKWithMeta(c, "daftar barang masuk berhasil diambil", list, utils.BuildPaginationMeta(p, total))
}

// Detail GET /barang-masuk/:id
func (h *Controller) Detail(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, msgIdBM, nil)
	}
	bm, err := h.repo.FindByID(id)
	if err != nil {
		return utils.Fail(c, fiber.StatusNotFound, constant.ErrBMTidakDitemukan, nil)
	}
	return utils.OK(c, "detail barang masuk berhasil diambil", bm)
}

// Create POST /barang-masuk — dibuat berstatus "draft"; stok & rak baru
// berubah setelah dokumen diselesaikan lewat Complete.
func (h *Controller) Create(c *fiber.Ctx) error {
	var req BMRequest
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
	if err := h.validateItems(req); err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	bm := &model.BarangMasuk{
		NomorPenerimaan: generateNomorBM(),
		PurchaseOrderID: req.PurchaseOrderID,
		SupplierID:      req.SupplierID,
		GudangID:        req.GudangID,
		Status:          constant.StatusBMDraft,
		Tanggal:         tanggal,
		Catatan:         req.Catatan,
		Items:           toItemModels(req.Items),
	}
	if err := h.repo.Create(bm); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal membuat dokumen barang masuk", nil)
	}
	return utils.Created(c, "dokumen barang masuk berhasil dibuat", bm)
}

func (h *Controller) requireDraft(id uint) (*model.BarangMasuk, error) {
	bm, err := h.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if bm.Status != constant.StatusBMDraft {
		return nil, errors.New(constant.ErrBMBukanDraft)
	}
	return bm, nil
}

// Update PUT /barang-masuk/:id — hanya boleh selama status masih draft.
func (h *Controller) Update(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, msgIdBM, nil)
	}
	bm, err := h.requireDraft(id)
	if err != nil {
		return utils.Fail(c, fiber.StatusConflict, err.Error(), nil)
	}

	var req BMRequest
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
	if err := h.validateItems(req); err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	bm.PurchaseOrderID = req.PurchaseOrderID
	bm.SupplierID = req.SupplierID
	bm.GudangID = req.GudangID
	bm.Tanggal = tanggal
	bm.Catatan = req.Catatan
	if err := h.repo.Update(bm, toItemModels(req.Items)); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal memperbarui dokumen barang masuk", nil)
	}
	return utils.OK(c, "dokumen barang masuk berhasil diperbarui", bm)
}

// Delete DELETE /barang-masuk/:id — hanya boleh selama status draft.
func (h *Controller) Delete(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, msgIdBM, nil)
	}
	if _, err := h.requireDraft(id); err != nil {
		return utils.Fail(c, fiber.StatusConflict, err.Error(), nil)
	}
	if err := h.repo.Delete(id); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal menghapus dokumen barang masuk", nil)
	}
	return utils.OK(c, "dokumen barang masuk berhasil dihapus", nil)
}

func (h *Controller) Complete(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, msgIdBM, nil)
	}
	userID, _ := c.Locals(constant.CtxUserID).(uint)

	bm, err := h.repo.Complete(id, userID)
	if err != nil {
		return utils.Fail(c, fiber.StatusConflict, err.Error(), nil)
	}
	return utils.OK(c, "barang masuk berhasil diselesaikan, stok & rak telah diperbarui", bm)
}

// Batalkan PATCH /barang-masuk/:id/batalkan
func (h *Controller) Batalkan(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, msgIdBM, nil)
	}
	bm, err := h.repo.Batalkan(id)
	if err != nil {
		return utils.Fail(c, fiber.StatusConflict, err.Error(), nil)
	}
	return utils.OK(c, "dokumen barang masuk berhasil dibatalkan", bm)
}

// Summary GET /barang-masuk/summary
func (h *Controller) Summary(c *fiber.Ctx) error {
	total, err := h.repo.CountByStatus("")
	draft, err2 := h.repo.CountByStatus(constant.StatusBMDraft)
	selesai, err3 := h.repo.CountByStatus(constant.StatusBMSelesai)
	if err != nil || err2 != nil || err3 != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal mengambil ringkasan", nil)
	}
	return utils.OK(c, "ringkasan barang masuk berhasil diambil", SummaryResponse{
		TotalDokumen: total, Draft: draft, Selesai: selesai,
	})
}

func (h *Controller) RegisterRoutes(router fiber.Router) {
	g := router.Group("/barang-masuk", middleware.JWTAuth(h.jwtSvc))

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
