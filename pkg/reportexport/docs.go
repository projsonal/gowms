package reportexport

import (
	"archive/zip"
	"bytes"
	"fmt"
	"html"
	"strings"
)

// ToDocx membuat file .docx minimal (kop surat + ringkasan + tabel) tanpa
// dependensi eksternal — format DOCX pada dasarnya adalah arsip ZIP berisi
// beberapa file XML (OOXML/WordprocessingML). Kita rakit langsung 3 bagian
// wajibnya (Content_Types, root relationship, dan word/document.xml)
// memakai archive/zip bawaan Go, jadi tidak perlu `go get` library docx
// pihak ketiga yang belum pasti bisa di-resolve di semua lingkungan build.
//
// Hasilnya memang lebih sederhana dari docx buatan Word asli (tanpa style
// kompleks), tapi valid dibuka Word/LibreOffice/Google Docs, dan cukup
// untuk kebutuhan laporan tabel seperti Excel/PDF di file lain paket ini.
func ToDocx(title string, summary [][2]string, headers []string, rows [][]string, chartInsight string) ([]byte, error) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	files := map[string]string{
		"[Content_Types].xml": contentTypesXML,
		"_rels/.rels":         rootRelsXML,
		"word/document.xml":   buildDocumentXML(title, summary, headers, rows, chartInsight),
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

func buildDocumentXML(title string, summary [][2]string, headers []string, rows [][]string, chartInsight string) string {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` + "\n")
	b.WriteString(`<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">`)
	b.WriteString(`<w:body>`)

	// Kop surat sederhana: nama aplikasi "WMS-RSD" (hijau tua, bold) lalu judul laporan.
	b.WriteString(`<w:p><w:pPr><w:jc w:val="center"/></w:pPr><w:r><w:rPr><w:b/><w:color w:val="146C14"/><w:sz w:val="36"/></w:rPr><w:t>WMS-RSD</w:t></w:r></w:p>`)
	b.WriteString(`<w:p><w:pPr><w:jc w:val="center"/></w:pPr><w:r><w:rPr><w:sz w:val="18"/></w:rPr><w:t>Warehouse Management System - RSD</w:t></w:r></w:p>`)
	b.WriteString(`<w:p/>`)

	// Judul laporan
	b.WriteString(`<w:p><w:pPr><w:jc w:val="center"/></w:pPr><w:r><w:rPr><w:b/><w:sz w:val="32"/></w:rPr><w:t>`)
	b.WriteString(escXML(title))
	b.WriteString(`</w:t></w:r></w:p>`)
	b.WriteString(`<w:p/>`) // baris kosong pemisah

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

	// Analisa Data — chart SUNGGUHAN tidak bisa disisipkan di format docx
	// ini (dirakit manual dari OOXML mentah, tanpa library yang bisa
	// men-generate/embed grafik) — sesuai instruksi eksplisit, dipakai
	// FALLBACK insight teks otomatis (angka yang sama dengan yang
	// divisualisasikan chart di web/PDF/Excel, cuma bentuknya kalimat).
	if chartInsight != "" {
		b.WriteString(`<w:p><w:r><w:rPr><w:b/></w:rPr><w:t>Analisa Data</w:t></w:r></w:p>`)
		b.WriteString(`<w:p><w:r><w:t xml:space="preserve">`)
		b.WriteString(escXML(chartInsight))
		b.WriteString(`</w:t></w:r></w:p>`)
		b.WriteString(`<w:p/>`)
	}

	// Tabel: header (bold) lalu tiap baris data.
	b.WriteString(`<w:tbl><w:tblPr><w:tblStyle w:val="TableGrid"/><w:tblW w:w="0" w:type="auto"/><w:tblBorders>`)
	for _, edge := range []string{"top", "left", "bottom", "right", "insideH", "insideV"} {
		b.WriteString(fmt.Sprintf(`<w:%s w:val="single" w:sz="4" w:space="0" w:color="999999"/>`, edge))
	}
	b.WriteString(`</w:tblBorders></w:tblPr>`)

	writeRow := func(cells []string, bold bool) {
		b.WriteString(`<w:tr>`)
		for _, cell := range cells {
			b.WriteString(`<w:tc><w:tcPr><w:tcW w:w="0" w:type="auto"/></w:tcPr><w:p><w:r>`)
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
		writeRow(headers, true)
	}
	for _, row := range rows {
		writeRow(row, false)
	}
	b.WriteString(`</w:tbl>`)
	// A4 potrait: lebar 11906 twip (≈210mm), tinggi 16838 twip (≈297mm) —
	// satuan resmi OOXML w:pgSz adalah twentieths of a point (twip).
	b.WriteString(`<w:sectPr><w:pgSz w:w="11906" w:h="16838" w:orient="portrait"/><w:pgMar w:top="1134" w:right="1134" w:bottom="1134" w:left="1134"/></w:sectPr>`)
	b.WriteString(`</w:body></w:document>`)
	return b.String()
}
