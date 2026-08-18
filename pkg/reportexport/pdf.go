package reportexport

import (
	"bytes"
	"strconv"
	"strings"
	"time"

	"github.com/jung-kurt/gofpdf"
)

// ChartData — hasil agregasi "Analisa Data" yang disisipkan ke file
// unduhan. Sengaja didefinisikan ULANG di sini (bukan import paket
// internal/controller/laporan) supaya pkg/reportexport tetap paket
// mandiri tanpa dependensi balik ke internal/ (pkg/ dirancang bisa
// dipakai controller manapun, tidak boleh bergantung ke satu controller
// spesifik).
type ChartData struct {
	Title  string
	Type   string // "bar" | "line" — saat ini keduanya digambar sebagai bar chart sederhana di PDF/Excel
	Labels []string
	Values []float64
}

// Kop surat WMS-RSD: pita hijau tua di atas & bawah tiap halaman, meniru
// gaya kop surat resmi perusahaan (lihat referensi Kopsurat_RSD.docx) tapi
// dengan nama aplikasi "WMS-RSD" alih-alih nama PT.
var (
	headerBandColor   = [3]int{20, 60, 20}  // hijau tua
	headerAccentColor = [3]int{90, 200, 60} // hijau terang (aksen garis diagonal)
	footerTextColor   = [3]int{255, 255, 255}
)

func drawLetterhead(pdf *gofpdf.Fpdf, reportTitle string) {
	pageWidth, _ := pdf.GetPageSize()

	// Pita header hijau tua penuh lebar halaman.
	pdf.SetFillColor(headerBandColor[0], headerBandColor[1], headerBandColor[2])
	pdf.Rect(0, 0, pageWidth, 20, "F")
	// Aksen garis diagonal khas kop surat referensi (disederhanakan jadi
	// beberapa garis miring hijau terang di ujung kanan pita).
	pdf.SetDrawColor(headerAccentColor[0], headerAccentColor[1], headerAccentColor[2])
	pdf.SetLineWidth(1.2)
	for i := 0; i < 3; i++ {
		x := pageWidth - 30 + float64(i)*6
		pdf.Line(x, 0, x-8, 20)
	}

	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "B", 16)
	pdf.SetXY(10, 4)
	pdf.CellFormat(pageWidth-20, 8, "WMS-RSD", "", 1, "L", false, 0, "")
	pdf.SetFont("Arial", "", 9)
	pdf.SetX(10)
	pdf.CellFormat(pageWidth-20, 6, "Warehouse Management System - RSD", "", 1, "L", false, 0, "")

	pdf.SetTextColor(0, 0, 0)
	pdf.SetY(26)
	pdf.SetFont("Arial", "B", 13)
	pdf.CellFormat(0, 8, reportTitle, "", 1, "L", false, 0, "")
	pdf.SetFont("Arial", "", 8)
	pdf.SetTextColor(90, 90, 90)
	pdf.CellFormat(0, 5, "Dicetak: "+time.Now().Format("2 January 2006 15:04")+" WIB", "", 1, "L", false, 0, "")
	pdf.SetTextColor(0, 0, 0)
	pdf.Ln(2)
}

func drawFooterBand(pdf *gofpdf.Fpdf) {
	pageWidth, pageHeight := pdf.GetPageSize()
	bandY := pageHeight - 14
	pdf.SetFillColor(headerBandColor[0], headerBandColor[1], headerBandColor[2])
	pdf.Rect(0, bandY, pageWidth, 14, "F")
	pdf.SetTextColor(footerTextColor[0], footerTextColor[1], footerTextColor[2])
	pdf.SetFont("Arial", "", 8)
	pdf.SetXY(10, bandY+4)
	pdf.CellFormat(pageWidth-20, 6, "www.wms-rsd.internal - Dokumen ini dihasilkan otomatis oleh sistem WMS-RSD", "", 0, "L", false, 0, "")
	pdf.SetTextColor(0, 0, 0)
}

// drawBarChart menggambar bar chart SEDERHANA langsung pakai primitif
// vektor gofpdf (Rect/Line/teks) — TIDAK ada library charting eksternal
// yang di-`go get` (proxy Go modules diblokir di sebagian lingkungan
// build/sandbox, jadi dependensi baru berisiko gagal fetch). Cukup untuk
// "Analisa Data" yang diminta (tren per periode / top item), bukan
// pengganti chart interaktif recharts yang dipakai di UI web.
//
// Ukuran chart MENYESUAIKAN OTOMATIS ke lebar halaman A4 potrait (bukan
// ukuran tetap) — barWidth dihitung dari lebar area usable dibagi jumlah
// titik data, supaya tetap proporsional baik untuk 3 titik (mis. laporan
// tahunan) maupun 31 titik (laporan harian sebulan penuh).
func drawBarChart(pdf *gofpdf.Fpdf, chart *ChartData) {
	if chart == nil || len(chart.Values) == 0 {
		return
	}
	pageWidth, _ := pdf.GetPageSize()
	marginL, _, marginR, _ := pdf.GetMargins()
	usable := pageWidth - marginL - marginR
	const chartHeight = 55.0 // mm — cukup besar untuk terbaca, tidak makan lebih dari ~1/4 halaman A4

	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(0, 7, "Analisa Data — "+chart.Title, "", 1, "L", false, 0, "")

	maxVal := 0.0
	for _, v := range chart.Values {
		if v > maxVal {
			maxVal = v
		}
	}
	if maxVal <= 0 {
		maxVal = 1
	}

	startY := pdf.GetY()
	startX := marginL
	// Sisakan ruang label sumbu-Y (angka maksimum) di kiri.
	yAxisLabelWidth := 12.0
	plotX := startX + yAxisLabelWidth
	plotWidth := usable - yAxisLabelWidth
	n := len(chart.Values)
	gap := 1.5
	barWidth := (plotWidth - gap*float64(n-1)) / float64(n)
	if barWidth < 2 {
		// Terlalu banyak titik data untuk lebar halaman — batasi ke 40
		// titik terakhir supaya chart tetap terbaca alih-alih bar setipis
		// rambut yang tidak berguna.
		const maxBars = 40
		if n > maxBars {
			chart.Labels = chart.Labels[n-maxBars:]
			chart.Values = chart.Values[n-maxBars:]
			n = maxBars
			barWidth = (plotWidth - gap*float64(n-1)) / float64(n)
		}
	}

	// Sumbu.
	pdf.SetDrawColor(180, 180, 180)
	pdf.Line(plotX, startY, plotX, startY+chartHeight)
	pdf.Line(plotX, startY+chartHeight, plotX+plotWidth, startY+chartHeight)
	pdf.SetFont("Arial", "", 6.5)
	pdf.SetXY(startX, startY-2)
	pdf.CellFormat(yAxisLabelWidth-1, 4, trimFloatPdf(maxVal), "", 0, "R", false, 0, "")

	pdf.SetFillColor(headerBandColor[0]+40, headerBandColor[1]+80, headerBandColor[2]+40)
	for i, v := range chart.Values {
		barH := (v / maxVal) * (chartHeight - 6)
		x := plotX + float64(i)*(barWidth+gap)
		y := startY + chartHeight - barH
		pdf.Rect(x, y, barWidth, barH, "F")
	}

	// Label sumbu-X — kalau kebanyakan titik, cuma tampilkan sebagian
	// (tiap-N) supaya tidak numpuk tak terbaca.
	pdf.SetFont("Arial", "", 5.5)
	pdf.SetTextColor(90, 90, 90)
	labelEvery := 1
	if n > 12 {
		labelEvery = (n + 11) / 12
	}
	for i, label := range chart.Labels {
		if i%labelEvery != 0 {
			continue
		}
		x := plotX + float64(i)*(barWidth+gap)
		pdf.SetXY(x-5, startY+chartHeight+1)
		pdf.CellFormat(barWidth+10, 4, label, "", 0, "C", false, 0, "")
	}
	pdf.SetTextColor(0, 0, 0)
	pdf.SetY(startY + chartHeight + 8)
}

func trimFloatPdf(f float64) string {
	s := strconv.FormatFloat(f, 'f', 1, 64)
	return strings.TrimSuffix(s, ".0")
}

// ToPDF membuat laporan PDF A4 POTRAIT dengan kop surat WMS-RSD di setiap
// halaman, ringkasan (summary [label,value]) sebelum tabel rincian — supaya
// data di luar tabel (total, agregat, dst — setara "chart data" di UI) ikut
// terbawa ke file yang diunduh, bukan cuma isi tabel rincian.
func ToPDF(title string, summary [][2]string, headers []string, rows [][]string, chart *ChartData) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetTitle(title, false)
	pdf.SetAutoPageBreak(true, 22)
	pdf.SetHeaderFunc(func() { drawLetterhead(pdf, title) })
	pdf.SetFooterFunc(func() { drawFooterBand(pdf) })
	pdf.AddPage()

	if len(summary) > 0 {
		pdf.SetFont("Arial", "B", 10)
		pdf.CellFormat(0, 7, "Ringkasan", "", 1, "L", false, 0, "")
		pdf.SetFont("Arial", "", 9)
		for _, kv := range summary {
			pdf.SetFont("Arial", "B", 9)
			pdf.CellFormat(45, 6, kv[0], "", 0, "L", false, 0, "")
			pdf.SetFont("Arial", "", 9)
			pdf.CellFormat(0, 6, kv[1], "", 1, "L", false, 0, "")
		}
		pdf.Ln(3)
	}

	drawBarChart(pdf, chart)

	pageWidth, _ := pdf.GetPageSize()
	marginL, _, marginR, _ := pdf.GetMargins()
	usable := pageWidth - marginL - marginR
	colWidth := usable
	if len(headers) > 0 {
		colWidth = usable / float64(len(headers))
	}

	pdf.SetFont("Arial", "B", 8)
	pdf.SetFillColor(230, 230, 230)
	for _, h := range headers {
		pdf.CellFormat(colWidth, 8, h, "1", 0, "C", true, 0, "")
	}
	pdf.Ln(-1)

	pdf.SetFont("Arial", "", 7.5)
	for _, row := range rows {
		for _, val := range row {
			pdf.CellFormat(colWidth, 7, val, "1", 0, "L", false, 0, "")
		}
		pdf.Ln(-1)
	}

	if pdf.Err() {
		return nil, pdf.Error()
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
