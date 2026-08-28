package dashboard

import (
	"bytes"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/xuri/excelize/v2"

	"github.com/projsonal/gowms/pkg/utils"
)

const dateFormat = "2006-01-02"
const prettyDateFormat = "02 Jan 2006"
const contentTypeHeader = "Content-Type"

func (h *Controller) ReportExport(c *fiber.Ctx) error {
	jenis := c.Params("jenis")
	format := c.Query("format", "pdf")
	fromStr := c.Query("from", "")
	toStr := c.Query("to", "")

	from, _ := parseDate(fromStr)
	to, _ := parseDate(toStr)
	if to.IsZero() {
		to = time.Now()
	}
	if from.IsZero() {
		from = to.AddDate(0, 0, -30)
	}
	to = time.Date(to.Year(), to.Month(), to.Day(), 23, 59, 59, 0, to.Location())

	type entry struct {
		Nomor   string
		Tanggal time.Time
		Info    string
		Status  string
		Qty     int64
	}
	var rows []entry
	db := h.db
	switch jenis {
	case "barang-masuk":
		db.Raw(`SELECT bm.nomor_penerimaan AS nomor, bm.tanggal, COALESCE(g.nama, '-') AS info, bm.status,
		         COALESCE((SELECT COUNT(*) FROM barang_masuk_items WHERE barang_masuk_id = bm.id), 0) AS qty
		         FROM barang_masuk bm LEFT JOIN gudangs g ON g.id = bm.gudang_id
		         WHERE bm.tanggal BETWEEN ? AND ? ORDER BY bm.tanggal DESC`, from, to).Scan(&rows)
	case "barang-keluar":
		db.Raw(`SELECT bk.nomor_pengeluaran AS nomor, bk.tanggal, COALESCE(g.nama, '-') AS info, bk.status,
		         COALESCE((SELECT COUNT(*) FROM barang_keluar_items WHERE barang_keluar_id = bk.id), 0) AS qty
		         FROM barang_keluar bk LEFT JOIN gudangs g ON g.id = bk.gudang_id
		         WHERE bk.tanggal BETWEEN ? AND ? ORDER BY bk.tanggal DESC`, from, to).Scan(&rows)
	case "stock-opname":
		db.Raw(`SELECT so.nomor_opname AS nomor, so.tanggal, COALESCE(g.nama, '-') AS info, so.status, 0 AS qty
		         FROM stock_opnames so LEFT JOIN gudangs g ON g.id = so.gudang_id
		         WHERE so.tanggal BETWEEN ? AND ? ORDER BY so.tanggal DESC`, from, to).Scan(&rows)
	case "stok":
		db.Raw(`SELECT b.kode_barang AS nomor, b.updated_at AS tanggal,
		         COALESCE(k.nama, '-') || ' - ' || b.nama AS info,
		         CASE WHEN b.stok <= b.stok_minimum THEN 'Kritis' ELSE 'Aman' END AS status,
		         b.stok AS qty FROM barangs b LEFT JOIN kategoris k ON k.id = b.kategori_id ORDER BY b.stok ASC`).Scan(&rows)
	default:
		return utils.Fail(c, fiber.StatusBadRequest, "jenis laporan tidak dikenal", nil)
	}
	if rows == nil {
		rows = []entry{}
	}

	title := prettyTitle(jenis)
	filename := fmt.Sprintf("%s_%s_%s", jenis, from.Format(dateFormat), to.Format(dateFormat))

	if format == "excel" || format == "xlsx" {
		f := excelize.NewFile()
		sheet := "Laporan"
		f.SetSheetName("Sheet1", sheet)
		f.SetCellValue(sheet, "A1", title)
		f.SetCellValue(sheet, "A2", fmt.Sprintf("Periode: %s s/d %s", from.Format(dateFormat), to.Format(dateFormat)))
		f.SetCellValue(sheet, "A3", fmt.Sprintf("Total: %d record", len(rows)))
		f.SetCellValue(sheet, "A5", "No.")
		f.SetCellValue(sheet, "B5", "Nomor")
		f.SetCellValue(sheet, "C5", "Tanggal")
		f.SetCellValue(sheet, "D5", "Info")
		f.SetCellValue(sheet, "E5", "Status")
		f.SetCellValue(sheet, "F5", "Qty/Total")

		for i, r := range rows {
			row := i + 6
			f.SetCellValue(sheet, cell("A", row), i+1)
			f.SetCellValue(sheet, cell("B", row), r.Nomor)
			f.SetCellValue(sheet, cell("C", row), r.Tanggal.Format("2006-01-02"))
			f.SetCellValue(sheet, cell("D", row), r.Info)
			f.SetCellValue(sheet, cell("E", row), r.Status)
			f.SetCellValue(sheet, cell("F", row), r.Qty)
		}
		for _, col := range []string{"A", "B", "C", "D", "E", "F"} {
			f.SetColWidth(sheet, col, col, 20)
		}
		var buf bytes.Buffer
		if err := f.Write(&buf); err != nil {
			return utils.Fail(c, fiber.StatusInternalServerError, "gagal membuat Excel: "+err.Error(), nil)
		}
		c.Set(contentTypeHeader, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.xlsx"`, filename))
		return c.Send(buf.Bytes())
	}

	shared := make([]struct {
		Nomor   string
		Tanggal time.Time
		Info    string
		Status  string
		Qty     int64
	}, len(rows))
	for i, r := range rows {
		shared[i] = struct {
			Nomor   string
			Tanggal time.Time
			Info    string
			Status  string
			Qty     int64
		}{r.Nomor, r.Tanggal, r.Info, r.Status, r.Qty}
	}
	html := renderReportHTML(title, from, to, shared)
	if format == "pdf" {
		c.Set(contentTypeHeader, "text/html; charset=utf-8")
		c.Set("Content-Disposition", fmt.Sprintf(`inline; filename="%s.html"`, filename))
		return c.SendString(html)
	}
	c.Set(contentTypeHeader, "text/html; charset=utf-8")
	return c.SendString(html)
}

func cell(col string, row int) string { return fmt.Sprintf("%s%d", col, row) }

func prettyTitle(jenis string) string {
	titles := map[string]string{
		"barang-masuk":  "Laporan Barang Masuk",
		"barang-keluar": "Laporan Barang Keluar",
		"stock-opname":  "Laporan Stock Opname",
		"stok":          "Laporan Stok",
	}
	if t, ok := titles[jenis]; ok {
		return t
	}
	return "Laporan"
}

func renderReportHTML(title string, from, to time.Time, rows []struct {
	Nomor   string
	Tanggal time.Time
	Info    string
	Status  string
	Qty     int64
}) string {
	var b bytes.Buffer
	b.WriteString(`<!DOCTYPE html><html lang="id"><head><meta charset="UTF-8">`)
	b.WriteString(`<title>`)
	b.WriteString(title)
	b.WriteString(`</title>`)
	b.WriteString(`<style>
	  * { box-sizing: border-box; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif; }
	  body { margin: 40px; color: #2b211d; }
	  header { border-bottom: 3px solid #b3471f; padding-bottom: 16px; margin-bottom: 24px; display: flex; align-items: center; justify-content: space-between; }
	  h1 { margin: 0; color: #3b140a; font-size: 24px; }
	  .brand { display: flex; align-items: center; gap: 12px; }
	  .brand-badge { background: linear-gradient(135deg, #3b140a, #5c2211); color: white; width: 48px; height: 48px; border-radius: 12px; display: grid; place-items: center; font-weight: bold; font-size: 20px; }
	  .meta { margin: 16px 0 24px; padding: 12px 16px; background: #fdf1ec; border-left: 4px solid #b3471f; border-radius: 6px; font-size: 13px; color: #8a7b74; }
	  .meta strong { color: #2b211d; }
	  table { width: 100%; border-collapse: collapse; margin-top: 8px; font-size: 12px; }
	  th { background: #3b140a; color: white; padding: 10px 12px; text-align: left; font-weight: 600; }
	  td { padding: 8px 12px; border-bottom: 1px solid #f0dad2; }
	  tr:nth-child(even) td { background: #fdf1ec; }
	  .status-selesai, .status-Terkirim, .status-Aman { color: #2f8132; font-weight: 600; }
	  .status-Kritis, .status-Gagal, .status-Batal { color: #c63c3c; font-weight: 600; }
	  .status-DalamPerjalanan, .status-Menunggu { color: #3454c7; font-weight: 600; }
	  footer { margin-top: 32px; padding-top: 16px; border-top: 1px solid #f0dad2; font-size: 10px; color: #8a7b74; text-align: center; }
	  @media print { body { margin: 20px; } header { page-break-inside: avoid; } .no-print { display: none; } }
	</style></head><body>`)
	b.WriteString(`<header><div class="brand"><div class="brand-badge">S</div><div><h1>`)
	b.WriteString(title)
	b.WriteString(`</h1><div style="font-size:12px;color:#8a7b74">StokRSD Warehouse Management System</div></div></div><div style="text-align:right;font-size:11px;color:#8a7b74">Dicetak: `)
	b.WriteString(time.Now().Format("02 Jan 2006 15:04"))
	b.WriteString(`</div></header>`)
	b.WriteString(fmt.Sprintf(`<div class="meta"><strong>Periode:</strong> %s s/d %s &nbsp;&middot;&nbsp; <strong>Total record:</strong> %d</div>`, from.Format(prettyDateFormat), to.Format(prettyDateFormat), len(rows)))
	b.WriteString(`<table><thead><tr><th style="width:40px">No</th><th>Nomor</th><th>Tanggal</th><th>Info</th><th>Status</th><th style="text-align:right">Qty/Total</th></tr></thead><tbody>`)
	for i, r := range rows {
		statusClass := "status-" + r.Status
		b.WriteString(fmt.Sprintf(`<tr><td>%d</td><td style="font-family:monospace;font-weight:600">%s</td><td>%s</td><td>%s</td><td class="%s">%s</td><td style="text-align:right">%d</td></tr>`,
			i+1,
			r.Nomor,
			r.Tanggal.Format(prettyDateFormat),
			r.Info,
			statusClass,
			r.Status,
			r.Qty))
	}
	if len(rows) == 0 {
		b.WriteString(`<tr><td colspan="6" style="text-align:center;padding:24px;color:#8a7b74">Tidak ada data pada periode ini</td></tr>`)
	}
	b.WriteString(`</tbody></table>`)
	b.WriteString(`<footer>Dokumen resmi StokRSD WMS &middot; www.stockrsd.local &middot; Digenerate otomatis oleh sistem</footer>`)
	b.WriteString(`<div class="no-print" style="position:fixed;bottom:20px;right:20px"><button onclick="window.print()" style="background:#b3471f;color:white;padding:10px 20px;border:none;border-radius:8px;font-weight:600;cursor:pointer">🖨️ Cetak / Simpan PDF</button></div>`)
	b.WriteString(`</body></html>`)
	return b.String()
}
