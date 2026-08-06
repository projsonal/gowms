package reportexport

import (
	"bytes"

	"github.com/jung-kurt/gofpdf"
)

func ToPDF(title string, headers []string, rows [][]string) ([]byte, error) {
	pdf := gofpdf.New("L", "mm", "A4", "")
	pdf.SetTitle(title, false)
	pdf.AddPage()

	pdf.SetFont("Arial", "B", 14)
	pdf.CellFormat(0, 10, title, "", 1, "L", false, 0, "")
	pdf.Ln(2)

	pageWidth, _ := pdf.GetPageSize()
	marginL, _, marginR, _ := pdf.GetMargins()
	usable := pageWidth - marginL - marginR
	colWidth := usable
	if len(headers) > 0 {
		colWidth = usable / float64(len(headers))
	}

	pdf.SetFont("Arial", "B", 9)
	pdf.SetFillColor(230, 230, 230)
	for _, h := range headers {
		pdf.CellFormat(colWidth, 8, h, "1", 0, "C", true, 0, "")
	}
	pdf.Ln(-1)

	pdf.SetFont("Arial", "", 8)
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
