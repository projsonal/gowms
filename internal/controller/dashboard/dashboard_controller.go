package dashboard

import (
	"github.com/gofiber/fiber/v2"

	"github.com/projsonal/gowms/internal/middleware"
	"github.com/projsonal/gowms/pkg/constant"
	"github.com/projsonal/gowms/pkg/utils"
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
	res.Gudang = GudangSummary{TotalGudang: totalGudang}

	bmDraft, _ := h.barangMasukRepo.CountByStatus(constant.StatusBMDraft)
	bmSelesai, _ := h.barangMasukRepo.CountByStatus(constant.StatusBMSelesai)
	res.BarangMasuk = DokumenSummary{Draft: bmDraft, Selesai: bmSelesai}

	bkDraft, _ := h.barangKeluarRepo.CountByStatus(constant.StatusBKDraft)
	bkSelesai, _ := h.barangKeluarRepo.CountByStatus(constant.StatusBKSelesai)
	res.BarangKeluar = DokumenSummary{Draft: bkDraft, Selesai: bkSelesai}

	soDraft, _ := h.stockOpnameRepo.CountByStatus(constant.StatusSODraft)
	soSelesai, _ := h.stockOpnameRepo.CountByStatus(constant.StatusSOSelesai)
	res.StockOpname = DokumenSummary{Draft: soDraft, Selesai: soSelesai}

	return utils.OK(c, "ringkasan dashboard berhasil diambil", res)
}

func (h *Controller) RegisterRoutes(router fiber.Router) {
	g := router.Group("/dashboard", middleware.JWTAuth(h.jwtSvc))
	view := middleware.RequirePermission(h.roleRepo, Module, constant.ActionView)
	g.Get("/summary", view, h.Summary)
	g.Get("/trend", view, h.Trend)
	g.Get("/activity", view, h.Activity)
	g.Get("/analisa", view, h.Analisa)
	g.Get("/notifications", view, h.Notifications)

	lg := router.Group("/laporan", middleware.JWTAuth(h.jwtSvc))
	lg.Get("/:jenis/preview", view, h.ReportPreview)
	lg.Get("/:jenis/export", view, h.ReportExport)
}
