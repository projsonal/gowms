package dashboard

import (
	"github.com/gofiber/fiber/v2"

	"github.com/projsonal/gostock/internal/middleware"
	"github.com/projsonal/gostock/pkg/constant"
	"github.com/projsonal/gostock/pkg/utils"
)

const Module = constant.ModuleDashboard

func (h *Controller) Summary(c *fiber.Ctx) error {
	var res DashboardResponse

	totalBarang, _ := h.barangRepo.CountAll()
	stokMenipis, _ := h.barangRepo.CountStokMenipis()
	nilaiInventaris, _ := h.barangRepo.SumNilaiInventaris()
	res.KelolaBarang = KelolaBarangSummary{
		TotalBarang: totalBarang, StokMenipis: stokMenipis, TotalNilaiInventaris: nilaiInventaris,
	}

	totalGudang, _ := h.gudangRepo.CountGudang()
	totalRak, _ := h.gudangRepo.CountRakAll()
	rakPenuh, _ := h.gudangRepo.CountRakByStatus("penuh")
	rakKosong, _ := h.gudangRepo.CountRakByStatus("kosong")
	res.Gudang = GudangSummary{TotalGudang: totalGudang, TotalRak: totalRak, RakPenuh: rakPenuh, RakKosong: rakKosong}

	totalSupplier, _ := h.supplierRepo.CountAll()
	supplierAktif, _ := h.supplierRepo.CountActive()
	res.Supplier = SupplierSummary{TotalSupplier: totalSupplier, SupplierAktif: supplierAktif}

	totalPO, _ := h.poRepo.CountByStatus("")
	menungguPO, _ := h.poRepo.CountByStatus(constant.StatusPODiajukan)
	disetujuiPO, _ := h.poRepo.CountByStatus(constant.StatusPODisetujui)
	res.PurchaseOrder = PurchaseOrderSummary{
		TotalPO: totalPO, MenungguPersetujuan: menungguPO, Disetujui: disetujuiPO,
	}

	bmDraft, _ := h.barangMasukRepo.CountByStatus(constant.StatusBMDraft)
	bmSelesai, _ := h.barangMasukRepo.CountByStatus(constant.StatusBMSelesai)
	res.BarangMasuk = DokumenSummary{Draft: bmDraft, Selesai: bmSelesai}

	bkDraft, _ := h.barangKeluarRepo.CountByStatus(constant.StatusBKDraft)
	bkSelesai, _ := h.barangKeluarRepo.CountByStatus(constant.StatusBKSelesai)
	res.BarangKeluar = DokumenSummary{Draft: bkDraft, Selesai: bkSelesai}

	soDraft, _ := h.stockOpnameRepo.CountByStatus(constant.StatusSODraft)
	soSelesai, _ := h.stockOpnameRepo.CountByStatus(constant.StatusSOSelesai)
	res.StockOpname = DokumenSummary{Draft: soDraft, Selesai: soSelesai}

	pgJalan, _ := h.pengirimanRepo.CountByStatus(constant.StatusPGDalamPerjalanan)
	pgTerkirim, _ := h.pengirimanRepo.CountByStatus(constant.StatusPGTerkirim)
	res.Pengiriman = PengirimanSummary{DalamPerjalanan: pgJalan, Terkirim: pgTerkirim}

	return utils.OK(c, "ringkasan dashboard berhasil diambil", res)
}

func (h *Controller) RegisterRoutes(router fiber.Router) {
	g := router.Group("/dashboard", middleware.JWTAuth(h.jwtSvc))
	view := middleware.RequirePermission(h.roleRepo, Module, constant.ActionView)
	g.Get("/summary", view, h.Summary)
	g.Get("/trend", view, h.Trend)
	g.Get("/activity", view, h.Activity)
	g.Get("/notifications", view, h.Notifications)
	g.Get("/courier-performance", view, h.CourierPerformance)
	g.Get("/gudang/beban", view, h.GudangBeban)

	lg := router.Group("/laporan", middleware.JWTAuth(h.jwtSvc))
	lg.Get("/:jenis/preview", view, h.ReportPreview)
	lg.Get("/:jenis/export", view, h.ReportExport)
}
