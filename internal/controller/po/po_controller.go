package purchase_order

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/projsonal/gowms/internal/middleware"
	"github.com/projsonal/gowms/internal/model"
	poRepo "github.com/projsonal/gowms/internal/repositories/po"
	"github.com/projsonal/gowms/pkg/constant"
	"github.com/projsonal/gowms/pkg/utils"
)

const Module = constant.ModulePurchaseOrder
const msgIdPo = "id purchase order tidak valid"

func parseIDParam(c *fiber.Ctx) (uint, error) {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

func generateNomorPO() string {
	return fmt.Sprintf("PO-%d-%d", time.Now().Year(), time.Now().UnixNano()%100000)
}

func toItemModels(items []ItemRequest) []model.PurchaseOrderItem {
	out := make([]model.PurchaseOrderItem, 0, len(items))
	for _, it := range items {
		out = append(out, model.PurchaseOrderItem{
			BarangID:    it.BarangID,
			QtyPesan:    it.QtyPesan,
			HargaSatuan: it.HargaSatuan,
		})
	}
	return out
}

// List GET /purchase-order?page=&limit=&search=&status=&supplier_id=
func (h *Controller) List(c *fiber.Ctx) error {
	p := utils.PaginationFromContext(c)
	supplierID, _ := strconv.ParseUint(c.Query("supplier_id", "0"), 10, 64)
	f := poRepo.Filter{Status: c.Query("status", ""), SupplierID: uint(supplierID)}

	list, total, err := h.repo.List(p, f)
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal mengambil daftar purchase order", nil)
	}
	return utils.OKWithMeta(c, "daftar purchase order berhasil diambil", list, utils.BuildPaginationMeta(p, total))
}

// Detail GET /purchase-order/:id
func (h *Controller) Detail(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, msgIdPo, nil)
	}
	po, err := h.repo.FindByID(id)
	if err != nil {
		return utils.Fail(c, fiber.StatusNotFound, constant.ErrPOTidakDitemukan, nil)
	}
	return utils.OK(c, "detail purchase order berhasil diambil", po)
}

func (h *Controller) Create(c *fiber.Ctx) error {
	var req PORequest
	if !utils.ParseAndValidate(c, &req) {
		return nil
	}
	if _, err := h.supplierRepo.FindByID(req.SupplierID); err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "supplier tidak ditemukan", nil)
	}

	po := &model.PurchaseOrder{
		NomorPO:          generateNomorPO(),
		SupplierID:       req.SupplierID,
		Status:           constant.StatusPODraft,
		TanggalPO:        req.TanggalPO,
		CatatanPengajuan: req.CatatanPengajuan,
		Items:            toItemModels(req.Items),
	}
	if err := h.repo.Create(po); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal membuat purchase order", nil)
	}
	return utils.Created(c, "purchase order berhasil dibuat", po)
}

func (h *Controller) requireDraft(id uint) (*model.PurchaseOrder, error) {
	po, err := h.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if po.Status != constant.StatusPODraft {
		return nil, errors.New(constant.ErrPOBukanDraft)
	}
	return po, nil
}

// Update PUT /purchase-order/:id — hanya boleh selama status masih draft.
func (h *Controller) Update(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, msgIdPo, nil)
	}
	po, err := h.requireDraft(id)
	if err != nil {
		return utils.Fail(c, fiber.StatusConflict, err.Error(), nil)
	}

	var req PORequest
	if !utils.ParseAndValidate(c, &req) {
		return nil
	}
	if _, err := h.supplierRepo.FindByID(req.SupplierID); err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "supplier tidak ditemukan", nil)
	}

	po.SupplierID = req.SupplierID
	po.TanggalPO = req.TanggalPO
	po.CatatanPengajuan = req.CatatanPengajuan
	if err := h.repo.Update(po, toItemModels(req.Items)); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal memperbarui purchase order", nil)
	}
	return utils.OK(c, "purchase order berhasil diperbarui", po)
}

// Delete DELETE /purchase-order/:id — hanya boleh selama status draft.
func (h *Controller) Delete(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, msgIdPo, nil)
	}
	if _, err := h.requireDraft(id); err != nil {
		return utils.Fail(c, fiber.StatusConflict, err.Error(), nil)
	}
	if err := h.repo.Delete(id); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal menghapus purchase order", nil)
	}
	return utils.OK(c, "purchase order berhasil dihapus", nil)
}

// Ajukan PATCH /purchase-order/:id/ajukan — draft -> diajukan.
func (h *Controller) Ajukan(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, msgIdPo, nil)
	}
	userID, _ := c.Locals(constant.CtxUserID).(uint)

	po, err := h.repo.Ajukan(id, userID)
	if err != nil {
		return utils.Fail(c, fiber.StatusConflict, constant.ErrPOBukanDraft, nil)
	}
	return utils.OK(c, "purchase order berhasil diajukan untuk persetujuan", po)
}

func (h *Controller) SetujuiTolak(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, msgIdPo, nil)
	}
	var req SetujuiTolakRequest
	if !utils.ParseAndValidate(c, &req) {
		return nil
	}
	userID, _ := c.Locals(constant.CtxUserID).(uint)

	po, err := h.repo.SetujuiTolak(id, userID, req.Setuju, req.Catatan)
	if err != nil {
		return utils.Fail(c, fiber.StatusConflict, err.Error(), nil)
	}
	msg := "purchase order berhasil ditolak"
	if req.Setuju {
		msg = "purchase order berhasil disetujui"
	}
	return utils.OK(c, msg, po)
}

// Batalkan PATCH /purchase-order/:id/batalkan
func (h *Controller) Batalkan(c *fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, msgIdPo, nil)
	}
	po, err := h.repo.Batalkan(id)
	if err != nil {
		return utils.Fail(c, fiber.StatusConflict, err.Error(), nil)
	}
	return utils.OK(c, "purchase order berhasil dibatalkan", po)
}

// Summary GET /purchase-order/summary — kartu ringkasan dashboard PO.
func (h *Controller) Summary(c *fiber.Ctx) error {
	total, err := h.repo.CountByStatus("")
	menunggu, err2 := h.repo.CountByStatus(constant.StatusPODiajukan)
	disetujui, err3 := h.repo.CountByStatus(constant.StatusPODisetujui)
	selesai, err4 := h.repo.CountByStatus(constant.StatusPOSelesai)
	if err != nil || err2 != nil || err3 != nil || err4 != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal mengambil ringkasan", nil)
	}
	return utils.OK(c, "ringkasan purchase order berhasil diambil", SummaryResponse{
		TotalPO: total, MenungguPersetujuan: menunggu, Disetujui: disetujui, Selesai: selesai,
	})
}

func (h *Controller) RegisterRoutes(router fiber.Router) {
	g := router.Group("/purchase-order", middleware.JWTAuth(h.jwtSvc))

	view := middleware.RequirePermission(h.roleRepo, Module, constant.ActionView)
	tambah := middleware.RequirePermission(h.roleRepo, Module, constant.ActionTambah)
	edit := middleware.RequirePermission(h.roleRepo, Module, constant.ActionEdit)
	approval := middleware.RequirePermission(h.roleRepo, Module, constant.ActionApprovalReject)

	g.Get("/summary", view, h.Summary)
	g.Get("/", view, h.List)
	g.Get("/:id", view, h.Detail)
	g.Post("/", tambah, h.Create)
	g.Put("/:id", edit, h.Update)
	g.Delete("/:id", edit, h.Delete)
	g.Patch("/:id/ajukan", edit, h.Ajukan)
	g.Patch("/:id/approval", approval, h.SetujuiTolak)
	g.Patch("/:id/batalkan", edit, h.Batalkan)
}
