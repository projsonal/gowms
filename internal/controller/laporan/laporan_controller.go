package laporan

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/projsonal/gowms/internal/middleware"
	barangRepoPkg "github.com/projsonal/gowms/internal/repositories/barang"
	barangKeluarRepoPkg "github.com/projsonal/gowms/internal/repositories/barang_keluar"
	barangMasukRepoPkg "github.com/projsonal/gowms/internal/repositories/barang_masuk"
	purchaseOrderRepoPkg "github.com/projsonal/gowms/internal/repositories/po"
	stockOpnameRepoPkg "github.com/projsonal/gowms/internal/repositories/stockOpname"
	"github.com/projsonal/gowms/pkg/constant"
	"github.com/projsonal/gowms/pkg/reportexport"
	"github.com/projsonal/gowms/pkg/utils"
)

const Module = constant.ModuleLaporan
const dateFormat = "2006-01-02"

func parseDateRange(c *fiber.Ctx) (dari, sampai *time.Time, err error) {
	if raw := c.Query("dari", ""); raw != "" {
		t, e := time.Parse(dateFormat, raw)
		if e != nil {
			return nil, nil, fmt.Errorf("format tanggal 'dari' tidak valid (gunakan YYYY-MM-DD)")
		}
		dari = &t
	}
	if raw := c.Query("sampai", ""); raw != "" {
		t, e := time.Parse(dateFormat, raw)
		if e != nil {
			return nil, nil, fmt.Errorf("format tanggal 'sampai' tidak valid (gunakan YYYY-MM-DD)")
		}
		t = t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		sampai = &t
	}
	return dari, sampai, nil
}

func inRange(t time.Time, dari, sampai *time.Time) bool {
	if dari != nil && t.Before(*dari) {
		return false
	}
	if sampai != nil && t.After(*sampai) {
		return false
	}
	return true
}

func formatRupiah(v int64) string {
	neg := v < 0
	s := strconv.FormatInt(v, 10)
	if neg {
		s = s[1:]
	}
	n := len(s)
	var parts []string
	for n > 3 {
		parts = append([]string{s[n-3:]}, parts...)
		s = s[:n-3]
		n = len(s)
	}
	parts = append([]string{s}, parts...)
	res := strings.Join(parts, ".")
	if neg {
		res = "-" + res
	}
	return res
}

func uintOrDash(v *uint) string {
	if v == nil {
		return "-"
	}
	return strconv.FormatUint(uint64(*v), 10)
}

func (h *Controller) buildStokBarang() (headers []string, rows [][]string, err error) {
	list, _, err := h.barangRepo.List(bigPagination(), barangRepoPkg.Filter{})
	if err != nil {
		return nil, nil, err
	}
	headers = []string{"Kode Barang", "Nama", "Kategori", "Satuan", "Stok", "Stok Minimum", "Harga Beli", "Nilai Inventaris", "Status"}
	for _, b := range list {
		kategori, satuan := "-", "-"
		if b.Kategori != nil {
			kategori = b.Kategori.Nama
		}
		if b.Satuan != nil {
			satuan = b.Satuan.Nama
		}
		status := "Aktif"
		if !b.IsActive {
			status = "Nonaktif"
		}
		rows = append(rows, []string{
			b.KodeBarang, b.Nama, kategori, satuan,
			strconv.Itoa(b.Stok), strconv.Itoa(b.StokMinimum),
			formatRupiah(b.HargaBeli), formatRupiah(b.NilaiInventaris()), status,
		})
	}
	return headers, rows, nil
}

func (h *Controller) buildBarangMasuk(dari, sampai *time.Time) (headers []string, rows [][]string, err error) {
	list, _, err := h.barangMasukRepo.List(bigPagination(), barangMasukRepoPkg.Filter{})
	if err != nil {
		return nil, nil, err
	}
	headers = []string{"Nomor Penerimaan", "Tanggal", "Gudang", "PO Terkait", "Status", "Diterima Oleh (User ID)", "Catatan"}
	for _, bm := range list {
		if !inRange(bm.Tanggal, dari, sampai) {
			continue
		}
		gudang, poNomor := "-", "-"
		if bm.Gudang != nil {
			gudang = bm.Gudang.Nama
		}
		if bm.PurchaseOrder != nil {
			poNomor = bm.PurchaseOrder.NomorPO
		}
		rows = append(rows, []string{
			bm.NomorPenerimaan, bm.Tanggal.Format(dateFormat), gudang, poNomor,
			bm.Status, uintOrDash(bm.DiterimaOleh), bm.Catatan,
		})
	}
	return headers, rows, nil
}

func (h *Controller) buildBarangKeluar(dari, sampai *time.Time) (headers []string, rows [][]string, err error) {
	list, _, err := h.barangKeluarRepo.List(bigPagination(), barangKeluarRepoPkg.Filter{})
	if err != nil {
		return nil, nil, err
	}
	headers = []string{"Nomor Pengeluaran", "Tanggal", "Gudang", "Keperluan", "Penerima", "Status", "Dikeluarkan Oleh (User ID)"}
	for _, bk := range list {
		if !inRange(bk.Tanggal, dari, sampai) {
			continue
		}
		gudang := "-"
		if bk.Gudang != nil {
			gudang = bk.Gudang.Nama
		}
		rows = append(rows, []string{
			bk.NomorPengeluaran, bk.Tanggal.Format(dateFormat), gudang, bk.Keperluan, bk.Penerima,
			bk.Status, uintOrDash(bk.DikeluarkanOleh),
		})
	}
	return headers, rows, nil
}

func (h *Controller) buildPurchaseOrder(dari, sampai *time.Time) (headers []string, rows [][]string, err error) {
	list, _, err := h.poRepo.List(bigPagination(), purchaseOrderRepoPkg.Filter{})
	if err != nil {
		return nil, nil, err
	}
	headers = []string{"Nomor PO", "Tanggal PO", "Supplier", "Status", "Total Estimasi"}
	for _, po := range list {
		if !inRange(po.TanggalPO, dari, sampai) {
			continue
		}
		supplier := "-"
		if po.Supplier != nil {
			supplier = po.Supplier.Nama
		}
		rows = append(rows, []string{
			po.NomorPO, po.TanggalPO.Format(dateFormat), supplier, po.Status, formatRupiah(po.TotalEstimasi),
		})
	}
	return headers, rows, nil
}

func (h *Controller) buildStockOpname(dari, sampai *time.Time) (headers []string, rows [][]string, err error) {
	list, _, err := h.stockOpnameRepo.List(bigPagination(), stockOpnameRepoPkg.Filter{})
	if err != nil {
		return nil, nil, err
	}
	headers = []string{"Nomor Opname", "Tanggal", "Gudang", "Status", "Dilakukan Oleh (User ID)", "Catatan"}
	for _, so := range list {
		if !inRange(so.Tanggal, dari, sampai) {
			continue
		}
		gudang := "-"
		if so.Gudang != nil {
			gudang = so.Gudang.Nama
		}
		rows = append(rows, []string{
			so.NomorOpname, so.Tanggal.Format(dateFormat), gudang, so.Status, uintOrDash(so.DilakukanOleh), so.Catatan,
		})
	}
	return headers, rows, nil
}

var reportTitles = map[string]string{
	constant.LaporanStokBarang:   "Laporan Stok Barang",
	constant.LaporanBarangMasuk:  "Laporan Barang Masuk",
	constant.LaporanBarangKeluar: "Laporan Barang Keluar",
	constant.LaporanPO:           "Laporan Purchase Order",
	constant.LaporanStokOpname:   "Laporan Stock Opname",
}

func (h *Controller) buildReport(tipe string, dari, sampai *time.Time) (title string, headers []string, rows [][]string, err error) {
	title, ok := reportTitles[tipe]
	if !ok {
		return "", nil, nil, errors.New(constant.ErrLaporanTipeTidakDidukung)
	}

	switch tipe {
	case constant.LaporanStokBarang:
		headers, rows, err = h.buildStokBarang()
	case constant.LaporanBarangMasuk:
		headers, rows, err = h.buildBarangMasuk(dari, sampai)
	case constant.LaporanBarangKeluar:
		headers, rows, err = h.buildBarangKeluar(dari, sampai)
	case constant.LaporanPO:
		headers, rows, err = h.buildPurchaseOrder(dari, sampai)
	case constant.LaporanStokOpname:
		headers, rows, err = h.buildStockOpname(dari, sampai)
	}
	return title, headers, rows, err
}

// computeGenericSummary membangun ringkasan generik (total baris + agregat
// kolom yang namanya mengindikasikan nilai uang/kuantitas) dari data
// laporan APA PUN — dipakai supaya file yang diunduh (Excel/PDF/Word) ikut
// membawa info "di luar tabel rincian" (setara kartu ringkasan/chart di
// UI), tanpa perlu logika ringkasan terpisah untuk tiap tipe laporan.
func computeGenericSummary(headers []string, rows [][]string) [][2]string {
	summary := [][2]string{{"Total Baris", strconv.Itoa(len(rows))}}

	for colIdx, header := range headers {
		lower := strings.ToLower(header)
		switch {
		case strings.Contains(lower, "nilai") || strings.Contains(lower, "harga") || strings.Contains(lower, "total"):
			var sum int64
			for _, row := range rows {
				if colIdx >= len(row) {
					continue
				}
				cleaned := strings.ReplaceAll(row[colIdx], ".", "")
				cleaned = strings.ReplaceAll(cleaned, ",", "")
				cleaned = strings.TrimSpace(strings.TrimPrefix(cleaned, "Rp"))
				if n, err := strconv.ParseInt(cleaned, 10, 64); err == nil {
					sum += n
				}
			}
			summary = append(summary, [2]string{"Total " + header, "Rp " + formatRupiah(sum)})
		case strings.Contains(lower, "stok") || strings.Contains(lower, "kuantitas") || strings.Contains(lower, "qty"):
			var sum int64
			for _, row := range rows {
				if colIdx >= len(row) {
					continue
				}
				if n, err := strconv.ParseInt(strings.TrimSpace(row[colIdx]), 10, 64); err == nil {
					sum += n
				}
			}
			summary = append(summary, [2]string{"Total " + header, strconv.FormatInt(sum, 10)})
		case strings.Contains(lower, "gudang"):
			distinct := map[string]struct{}{}
			for _, row := range rows {
				if colIdx < len(row) && row[colIdx] != "" && row[colIdx] != "-" {
					distinct[row[colIdx]] = struct{}{}
				}
			}
			if len(distinct) > 0 {
				summary = append(summary, [2]string{"Gudang Terlibat", strconv.Itoa(len(distinct))})
			}
		}
	}
	return summary
}

func (h *Controller) Export(c *fiber.Ctx) error {
	tipe := c.Query("tipe", "")
	format := c.Query("format", "")
	if format != constant.FormatExcel && format != constant.FormatPDF && format != constant.FormatWord {
		return utils.Fail(c, fiber.StatusBadRequest, constant.ErrLaporanFormatTidakDidukung, nil)
	}

	dari, sampai, err := parseDateRange(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	title, headers, rows, err := h.buildReport(tipe, dari, sampai)
	if err != nil {
		status := fiber.StatusInternalServerError
		if err.Error() == constant.ErrLaporanTipeTidakDidukung {
			status = fiber.StatusBadRequest
		}
		return utils.Fail(c, status, err.Error(), nil)
	}

	timestamp := time.Now().Format("20060102-150405")
	summary := computeGenericSummary(headers, rows)
	switch format {
	case constant.FormatExcel:
		data, err := reportexport.ToExcel(title, summary, headers, rows)
		if err != nil {
			return utils.Fail(c, fiber.StatusInternalServerError, "gagal membuat file excel", nil)
		}
		c.Set(fiber.HeaderContentType, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		c.Set(fiber.HeaderContentDisposition, fmt.Sprintf(`attachment; filename="%s-%s.xlsx"`, tipe, timestamp))
		return c.Send(data)
	case constant.FormatPDF:
		data, err := reportexport.ToPDF(title, summary, headers, rows)
		if err != nil {
			return utils.Fail(c, fiber.StatusInternalServerError, "gagal membuat file pdf", nil)
		}
		c.Set(fiber.HeaderContentType, "application/pdf")
		c.Set(fiber.HeaderContentDisposition, fmt.Sprintf(`attachment; filename="%s-%s.pdf"`, tipe, timestamp))
		return c.Send(data)
	case constant.FormatWord:
		data, err := reportexport.ToDocx(title, summary, headers, rows)
		if err != nil {
			return utils.Fail(c, fiber.StatusInternalServerError, "gagal membuat file docx", nil)
		}
		c.Set(fiber.HeaderContentType, "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
		c.Set(fiber.HeaderContentDisposition, fmt.Sprintf(`attachment; filename="%s-%s.docx"`, tipe, timestamp))
		return c.Send(data)
	}
	return utils.Fail(c, fiber.StatusBadRequest, constant.ErrLaporanFormatTidakDidukung, nil)
}

func (h *Controller) Types(c *fiber.Ctx) error {
	types := make([]fiber.Map, 0, len(reportTitles))
	for key, label := range reportTitles {
		types = append(types, fiber.Map{"tipe": key, "label": label})
	}
	return utils.OK(c, "daftar tipe laporan berhasil diambil", types)
}

// Preview GET /laporan/preview?tipe=&dari=&sampai= — dipakai halaman
// laporan di frontend untuk menampilkan tabel "Rincian Laporan" & kartu
// ringkasan dengan data ASLI dari database, sebelum user memutuskan mau
// diunduh atau tidak (sebelumnya frontend cuma menampilkan data dummy dan
// baru memanggil backend saat tombol "Unduh Laporan" ditekan).
func (h *Controller) Preview(c *fiber.Ctx) error {
	tipe := c.Query("tipe", "")
	dari, sampai, err := parseDateRange(c)
	if err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, err.Error(), nil)
	}
	title, headers, rows, err := h.buildReport(tipe, dari, sampai)
	if err != nil {
		status := fiber.StatusInternalServerError
		if err.Error() == constant.ErrLaporanTipeTidakDidukung {
			status = fiber.StatusBadRequest
		}
		return utils.Fail(c, status, err.Error(), nil)
	}
	summaryPairs := computeGenericSummary(headers, rows)
	summary := make([]fiber.Map, 0, len(summaryPairs))
	for _, kv := range summaryPairs {
		summary = append(summary, fiber.Map{"label": kv[0], "value": kv[1]})
	}
	return utils.OK(c, "pratinjau laporan berhasil diambil", fiber.Map{
		"title":   title,
		"headers": headers,
		"rows":    rows,
		"summary": summary,
	})
}

// RegisterRoutes mendaftarkan endpoint modul "Laporan".
func (h *Controller) RegisterRoutes(router fiber.Router) {
	g := router.Group("/laporan", middleware.JWTAuth(h.jwtSvc))
	view := middleware.RequirePermission(h.roleRepo, Module, constant.ActionView)
	print := middleware.RequirePermission(h.roleRepo, Module, constant.ActionPrint)

	g.Get("/tipe", view, h.Types)
	g.Get("/preview", view, h.Preview)
	g.Get("/export", print, h.Export)
}
