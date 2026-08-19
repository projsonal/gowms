package reportexport

import (
	"archive/zip"
	"bytes"
	"fmt"
	"html"
	"strconv"
	"strings"
)

// ToDocx membuat file .docx (kop surat + ringkasan + "Analisa Data" +
// tabel) tanpa dependensi eksternal — format DOCX pada dasarnya adalah
// arsip ZIP berisi beberapa file XML (OOXML/WordprocessingML). Kita
// rakit langsung 3 bagian wajibnya (Content_Types, root relationship,
// dan word/document.xml) memakai archive/zip bawaan Go, jadi tidak perlu
// `go get` library docx pihak ketiga yang belum pasti bisa di-resolve di
// semua lingkungan build.
//
// Tampilan disamakan dengan pkg/reportexport/pdf.go (pita hijau tua di
// atas berisi "WMS-RSD", judul laporan rata kiri, tabel header abu-abu,
// bar chart "Analisa Data") — SESUAI PERMINTAAN EKSPLISIT supaya
// unduhan PDF & Word terlihat konsisten. `chartInsight` tetap diterima
// terpisah dari `chart` (bukan dihitung ulang di sini) karena rumus
// ringkasannya (total/rata-rata/tertinggi/terendah) sudah ada di
// internal/controller/laporan (ChartData.Insight()) — reportexport
// sengaja tidak boleh bergantung balik ke internal/, lihat catatan di
// pdf.go soal ChartData didefinisikan ulang.
//
// SATU KETERBATASAN JUJUR: pita hijau di BAWAH (footer berulang tiap
// halaman) di PDF TIDAK direplikasi persis di sini — footer berulang
// asli OOXML butuh bagian file terpisah (word/footer1.xml + relationship
// + rujukan di sectPr) yang menambah kerumitan & risiko file korup kalau
// salah satu potongannya meleset, sedangkan bagian ini dirakit manual
// tanpa alat bantu Word untuk validasi. Sebagai gantinya, tagline yang
// sama ditulis sebagai paragraf penutup biasa di akhir dokumen — tetap
// ada infonya, cuma tidak berulang di tiap halaman seperti PDF.
func ToDocx(title string, summary [][2]string, headers []string, rows [][]string, chart *ChartData, chartInsight string) ([]byte, error) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	files := map[string]string{
		"[Content_Types].xml": contentTypesXML,
		"_rels/.rels":         rootRelsXML,
		"word/document.xml":   buildDocumentXML(title, summary, headers, rows, chart, chartInsight),
	}

	// Urutan penulisan disamakan dengan urutan map di atas supaya
	// deterministik (peta Go tidak berurutan) — beberapa reader docx
	// (terutama versi lama) lebih toleran kalau [Content_Types].xml
	// berada di awal arsip.
	order := []string{"[Content_Types].xml", "_rels/.rels", "word/document.xml"}
	for _, name := range order {
		w, err := zw.Create(name)
		if err != nil {
			return nil, fmt.Errorf("reportexport: gagal membuat entri %s: %w", name, err)
		}
		if _, err := w.Write([]byte(files[name])); err != nil {
			return nil, fmt.Errorf("reportexport: gagal menulis entri %s: %w", name, err)
		}
	}

	if err := zw.Close(); err != nil {
		return nil, fmt.Errorf("reportexport: gagal finalisasi docx: %w", err)
	}
	return buf.Bytes(), nil
}

const contentTypesXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
</Types>`

const rootRelsXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>
</Relationships>`

func escXML(s string) string {
	return html.EscapeString(s)
}

// docxPageWidthTwip — lebar usable A4 potrait dikurangi margin kiri+kanan
// (masing-masing 1134 twip, lihat w:sectPr di bawah), dipakai supaya
// pita banner & tabel bar chart selebar penuh halaman, konsisten dengan
// lebar usable di pdf.go.
const docxPageWidthTwip = 9638

func buildDocumentXML(title string, summary [][2]string, headers []string, rows [][]string, chart *ChartData, chartInsight string) string {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` + "\n")
	b.WriteString(`<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">`)
	b.WriteString(`<w:body>`)

	writeBannerHeader(&b)
	writeTitleAndPrinted(&b, title)
	writeSummary(&b, summary)
	writeAnalysisSection(&b, chart, chartInsight)
	writeDataTable(&b, headers, rows)
	writeClosingTagline(&b)

	// A4 potrait: lebar 11906 twip (≈210mm), tinggi 16838 twip (≈297mm) —
	// satuan resmi OOXML w:pgSz adalah twentieths of a point (twip).
	b.WriteString(`<w:sectPr><w:pgSz w:w="11906" w:h="16838" w:orient="portrait"/><w:pgMar w:top="1134" w:right="1134" w:bottom="1134" w:left="1134"/></w:sectPr>`)
	b.WriteString(`</w:body></w:document>`)
	return b.String()
}

func writeTitleAndPrinted(b *strings.Builder, title string) {
	// Judul laporan + "Dicetak" — rata KIRI (bukan tengah), teks hitam
	// biasa, meniru posisi & gaya di drawLetterhead (pdf.go).
	b.WriteString(`<w:p><w:pPr><w:spacing w:before="120"/></w:pPr><w:r><w:rPr><w:b/><w:sz w:val="28"/></w:rPr><w:t>`)
	b.WriteString(escXML(title))
	b.WriteString(`</w:t></w:r></w:p>`)
	b.WriteString(`<w:p><w:r><w:rPr><w:color w:val="5A5A5A"/><w:sz w:val="16"/></w:rPr><w:t>Dicetak otomatis oleh sistem WMS-RSD</w:t></w:r></w:p>`)
	b.WriteString(`<w:p/>`)
}

func writeSummary(b *strings.Builder, summary [][2]string) {
	// Ringkasan (data di luar tabel rincian — total, agregat, dst.)
	if len(summary) > 0 {
		b.WriteString(`<w:p><w:r><w:rPr><w:b/></w:rPr><w:t>Ringkasan</w:t></w:r></w:p>`)
		for _, kv := range summary {
			b.WriteString(`<w:p><w:r><w:rPr><w:b/></w:rPr><w:t xml:space="preserve">`)
			b.WriteString(escXML(kv[0]))
			b.WriteString(`: </w:t></w:r><w:r><w:t xml:space="preserve">`)
			b.WriteString(escXML(kv[1]))
			b.WriteString(`</w:t></w:r></w:p>`)
		}
		b.WriteString(`<w:p/>`)
	}
}

func writeAnalysisSection(b *strings.Builder, chart *ChartData, chartInsight string) {
	// "Analisa Data — {judul chart}" — SAMA PERSIS teks headingnya dengan
	// drawBarChart di pdf.go. Chart sungguhan (gambar vektor) tidak bisa
	// disisipkan di sini (dirakit manual dari OOXML mentah, tanpa library
	// chart apa pun), jadi diganti bar chart TEKS memakai tabel dengan sel
	// diarsir (lihat writeBarChartTable) — supaya tetap ada representasi
	// VISUAL proporsional, bukan cuma kalimat. Kalimat insight otomatis
	// (total/rata-rata/tertinggi/terendah) tetap ditulis di bawahnya
	// sebagai pelengkap.
	if chart != nil && len(chart.Values) > 0 {
		b.WriteString(`<w:p><w:r><w:rPr><w:b/></w:rPr><w:t xml:space="preserve">Analisa Data — </w:t></w:r><w:r><w:rPr><w:b/></w:rPr><w:t>`)
		b.WriteString(escXML(chart.Title))
		b.WriteString(`</w:t></w:r></w:p>`)
		writeBarChartTable(b, chart)
		if chartInsight != "" {
			b.WriteString(`<w:p><w:pPr><w:spacing w:before="80"/></w:pPr><w:r><w:rPr><w:i/><w:sz w:val="18"/></w:rPr><w:t xml:space="preserve">`)
			b.WriteString(escXML(chartInsight))
			b.WriteString(`</w:t></w:r></w:p>`)
		}
		b.WriteString(`<w:p/>`)
	} else if chartInsight != "" {
		// Tidak ada titik data sama sekali (mis. periode kosong) — tetap
		// tampilkan kalimat insight-nya saja (biasanya "Belum ada data
		// yang cukup...", lihat ChartData.Insight()).
		b.WriteString(`<w:p><w:r><w:rPr><w:b/></w:rPr><w:t>Analisa Data</w:t></w:r></w:p>`)
		b.WriteString(`<w:p><w:r><w:t xml:space="preserve">`)
		b.WriteString(escXML(chartInsight))
		b.WriteString(`</w:t></w:r></w:p>`)
		b.WriteString(`<w:p/>`)
	}
}

func writeDataTable(b *strings.Builder, headers []string, rows [][]string) {
	// Tabel: header (bold, diarsir abu-abu — samakan dengan header abu-abu
	// di pdf.go) lalu tiap baris data.
	b.WriteString(`<w:tbl><w:tblPr><w:tblStyle w:val="TableGrid"/><w:tblW w:w="0" w:type="auto"/><w:tblBorders>`)
	for _, edge := range []string{"top", "left", "bottom", "right", "insideH", "insideV"} {
		b.WriteString(fmt.Sprintf(`<w:%s w:val="single" w:sz="4" w:space="0" w:color="999999"/>`, edge))
	}
	b.WriteString(`</w:tblBorders></w:tblPr>`)

	writeRow := func(cells []string, bold bool, shadeFill string) {
		b.WriteString(`<w:tr>`)
		for _, cell := range cells {
			b.WriteString(`<w:tc><w:tcPr><w:tcW w:w="0" w:type="auto"/>`)
			if shadeFill != "" {
				b.WriteString(fmt.Sprintf(`<w:shd w:val="clear" w:fill="%s"/>`, shadeFill))
			}
			b.WriteString(`</w:tcPr><w:p><w:r>`)
			if bold {
				b.WriteString(`<w:rPr><w:b/></w:rPr>`)
			}
			b.WriteString(`<w:t xml:space="preserve">`)
			b.WriteString(escXML(cell))
			b.WriteString(`</w:t></w:r></w:p></w:tc>`)
		}
		b.WriteString(`</w:tr>`)
	}

	if len(headers) > 0 {
		writeRow(headers, true, "E6E6E6")
	}
	for _, row := range rows {
		writeRow(row, false, "")
	}
	b.WriteString(`</w:tbl>`)
	b.WriteString(`<w:p/>`)
}

func writeClosingTagline(b *strings.Builder) {
	// Tagline penutup — lihat catatan keterbatasan di komentar ToDocx:
	// ini BUKAN footer berulang tiap halaman seperti di PDF, cuma
	// paragraf biasa di akhir dokumen.
	b.WriteString(`<w:p><w:pPr><w:jc w:val="center"/></w:pPr><w:r><w:rPr><w:color w:val="146C14"/><w:sz w:val="16"/></w:rPr><w:t>www.wms-rsd.internal - Dokumen ini dihasilkan otomatis oleh sistem WMS-RSD</w:t></w:r></w:p>`)
}

// writeBannerHeader menulis pita hijau tua selebar halaman berisi
// "WMS-RSD" (putih, tebal) + tagline — pendekatan OOXML untuk latar
// belakang berwarna adalah tabel satu sel dengan w:shd (shading), karena
// paragraf biasa tidak punya properti latar belakang selebar halaman.
// Meniru drawLetterhead (pdf.go) sedekat mungkin dalam batas kemampuan
// docx yang dirakit manual (tanpa DrawingML untuk aksen garis diagonal).
func writeBannerHeader(b *strings.Builder) {
	fmt.Fprintf(b, `<w:tbl><w:tblPr><w:tblW w:w="%d" w:type="dxa"/><w:tblBorders>`, docxPageWidthTwip)
	for _, edge := range []string{"top", "left", "bottom", "right", "insideH", "insideV"} {
		fmt.Fprintf(b, `<w:%s w:val="none" w:sz="0" w:space="0" w:color="auto"/>`, edge)
	}
	b.WriteString(`</w:tblBorders></w:tblPr>`)
	fmt.Fprintf(b, `<w:tblGrid><w:gridCol w:w="%d"/></w:tblGrid>`, docxPageWidthTwip)
	b.WriteString(`<w:tr><w:tc><w:tcPr>`)
	fmt.Fprintf(b, `<w:tcW w:w="%d" w:type="dxa"/>`, docxPageWidthTwip)
	b.WriteString(`<w:shd w:val="clear" w:fill="143C14"/><w:tcMar><w:top w:w="120" w:type="dxa"/><w:bottom w:w="120" w:type="dxa"/><w:left w:w="120" w:type="dxa"/></w:tcMar></w:tcPr>`)
	b.WriteString(`<w:p><w:r><w:rPr><w:b/><w:color w:val="FFFFFF"/><w:sz w:val="32"/></w:rPr><w:t>WMS-RSD</w:t></w:r></w:p>`)
	b.WriteString(`<w:p><w:r><w:rPr><w:color w:val="FFFFFF"/><w:sz w:val="18"/></w:rPr><w:t>Warehouse Management System - RSD</w:t></w:r></w:p>`)
	b.WriteString(`</w:tc></w:tr></w:tbl>`)
}

// writeBarChartTable menggambar bar chart TEKS pakai tabel: satu baris
// per titik data, kolom kiri = label, kolom kanan = DUA sel bersebelahan
// (bar terisi diarsir hijau + sisa kosong) yang lebarnya proporsional ke
// nilai/nilai-maksimum — pendekatan paling dekat ke bar chart visual
// sungguhan yang bisa dicapai OOXML mentah tanpa DrawingML/gambar. Sama
// konsepnya dengan drawBarChart di pdf.go, cuma horizontal (per baris)
// bukan vertikal (per kolom) karena tabel Word jauh lebih gampang diatur
// lebar KOLOM-nya daripada tinggi baris presisi.
func writeBarChartTable(b *strings.Builder, chart *ChartData) {
	const labelColTwip = 2200
	const barTrackTwip = 5000 // lebar total "trek" bar (terisi + kosong)
	const valueColTwip = docxPageWidthTwip - labelColTwip - barTrackTwip

	maxVal := 0.0
	for _, v := range chart.Values {
		if v > maxVal {
			maxVal = v
		}
	}
	if maxVal <= 0 {
		maxVal = 1
	}

	b.WriteString(`<w:tbl><w:tblPr><w:tblW w:w="0" w:type="auto"/><w:tblBorders>`)
	for _, edge := range []string{"top", "left", "bottom", "right", "insideH", "insideV"} {
		fmt.Fprintf(b, `<w:%s w:val="none" w:sz="0" w:space="0" w:color="auto"/>`, edge)
	}
	b.WriteString(`</w:tblBorders></w:tblPr>`)
	fmt.Fprintf(b, `<w:tblGrid><w:gridCol w:w="%d"/><w:gridCol w:w="%d"/><w:gridCol w:w="%d"/><w:gridCol w:w="%d"/></w:tblGrid>`,
		labelColTwip, barTrackTwip/2, barTrackTwip/2, valueColTwip)

	// Batasi jumlah baris ditampilkan (konsisten dengan pdf.go yang juga
	// membatasi ke 40 titik terakhir kalau kebanyakan) — di docx kita
	// lebih ketat (20) karena tiap baris makan ruang vertikal jauh lebih
	// banyak daripada bar tipis di PDF.
	const maxRows = 20
	labels, values := chart.Labels, chart.Values
	if len(values) > maxRows {
		labels = labels[len(labels)-maxRows:]
		values = values[len(values)-maxRows:]
	}

	for i, v := range values {
		filledRatio := v / maxVal
		filledTwip := int(filledRatio * float64(barTrackTwip))
		if filledTwip < 0 {
			filledTwip = 0
		}
		if filledTwip > barTrackTwip {
			filledTwip = barTrackTwip
		}
		emptyTwip := barTrackTwip - filledTwip

		label := ""
		if i < len(labels) {
			label = labels[i]
		}

		b.WriteString(`<w:tr>`)
		fmt.Fprintf(b, `<w:tc><w:tcPr><w:tcW w:w="%d" w:type="dxa"/></w:tcPr><w:p><w:r><w:rPr><w:sz w:val="16"/></w:rPr><w:t xml:space="preserve">`, labelColTwip)
		b.WriteString(escXML(label))
		b.WriteString(`</w:t></w:r></w:p></w:tc>`)

		fmt.Fprintf(b, `<w:tc><w:tcPr><w:tcW w:w="%d" w:type="dxa"/><w:shd w:val="clear" w:fill="5AC83C"/></w:tcPr><w:p/></w:tc>`, filledTwip)
		fmt.Fprintf(b, `<w:tc><w:tcPr><w:tcW w:w="%d" w:type="dxa"/></w:tcPr><w:p/></w:tc>`, emptyTwip)

		fmt.Fprintf(b, `<w:tc><w:tcPr><w:tcW w:w="%d" w:type="dxa"/></w:tcPr><w:p><w:r><w:rPr><w:sz w:val="16"/></w:rPr><w:t xml:space="preserve">`, valueColTwip)
		b.WriteString(escXML(trimFloatDocx(v)))
		b.WriteString(`</w:t></w:r></w:p></w:tc>`)
		b.WriteString(`</w:tr>`)
	}
	b.WriteString(`</w:tbl>`)
}

func trimFloatDocx(f float64) string {
	s := strconv.FormatFloat(f, 'f', 1, 64)
	return strings.TrimSuffix(s, ".0")
}