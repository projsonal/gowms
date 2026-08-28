package reportexport

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

const maxSheetNameLen = 31

func sanitizeSheetName(title string) string {
	if title == "" {
		return "Laporan"
	}
	if len(title) > maxSheetNameLen {
		return title[:maxSheetNameLen]
	}
	return title
}

type aggregateKind int

const (
	aggregateNone aggregateKind = iota
	aggregateCurrency
	aggregateQty
	aggregateGudang
)

func classifyColumn(header string) aggregateKind {
	lower := strings.ToLower(header)
	switch {
	case strings.Contains(lower, "nilai") || strings.Contains(lower, "harga") || strings.Contains(lower, "total"):
		return aggregateCurrency
	case strings.Contains(lower, "stok") || strings.Contains(lower, "kuantitas") || strings.Contains(lower, "qty"):
		return aggregateQty
	case strings.Contains(lower, "gudang"):
		return aggregateGudang
	default:
		return aggregateNone
	}
}

func parseNumericCell(val string) (float64, bool) {
	cleaned := strings.TrimSpace(val)
	cleaned = strings.TrimPrefix(cleaned, "Rp")
	cleaned = strings.TrimSpace(cleaned)
	cleaned = strings.ReplaceAll(cleaned, ".", "")
	cleaned = strings.ReplaceAll(cleaned, ",", "")
	if cleaned == "" || cleaned == "-" {
		return 0, false
	}
	n, err := strconv.ParseFloat(cleaned, 64)
	if err != nil {
		return 0, false
	}
	return n, true
}

func ToExcel(title string, summary [][2]string, headers []string, rows [][]string, chart *ChartData) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	sheet := sanitizeSheetName(title)
	if err := f.SetSheetName(f.GetSheetName(0), sheet); err != nil {
		return nil, err
	}

	boldStyle, err := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})
	if err != nil {
		return nil, err
	}
	titleStyle, err := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true, Size: 14, Color: "146C14"}})
	if err != nil {
		return nil, err
	}
	currencyStyle, err := f.NewStyle(&excelize.Style{
		Font:         &excelize.Font{},
		CustomNumFmt: strPtr(`"Rp"\ #,##0`),
	})
	if err != nil {
		return nil, err
	}
	qtyStyle, err := f.NewStyle(&excelize.Style{
		Font:         &excelize.Font{},
		CustomNumFmt: strPtr(`#,##0`),
	})
	if err != nil {
		return nil, err
	}
	boldCurrencyStyle, err := f.NewStyle(&excelize.Style{
		Font:         &excelize.Font{Bold: true},
		CustomNumFmt: strPtr(`"Rp"\ #,##0`),
	})
	if err != nil {
		return nil, err
	}
	boldQtyStyle, err := f.NewStyle(&excelize.Style{
		Font:         &excelize.Font{Bold: true},
		CustomNumFmt: strPtr(`#,##0`),
	})
	if err != nil {
		return nil, err
	}

	row, err := writeHeader(f, sheet, title, titleStyle)
	if err != nil {
		return nil, err
	}

	summaryRows := buildSummaryRows(headers)
	tableHeaderRow := row
	if len(summaryRows) > 0 {
		tableHeaderRow = row + 1 + len(summaryRows) + 1
	}
	dataStartRow := tableHeaderRow + 1
	dataEndRow := dataStartRow + len(rows) - 1
	if len(rows) == 0 {
		dataEndRow = dataStartRow
	}

	if err := writeSummaryFormulas(f, sheet, summaryRows, row, summaryStyles{
		bold:         boldStyle,
		boldCurrency: boldCurrencyStyle,
		boldQty:      boldQtyStyle,
	}, dataStartRow, dataEndRow); err != nil {
		return nil, err
	}

	if err := writeHeaders(f, sheet, headers, tableHeaderRow, boldStyle); err != nil {
		return nil, err
	}

	colKinds := make([]aggregateKind, len(headers))
	for i, h := range headers {
		colKinds[i] = classifyColumn(h)
	}

	if err := writeRows(f, sheet, rows, tableHeaderRow, colKinds, rowCellStyles{
		currency: currencyStyle,
		qty:      qtyStyle,
	}); err != nil {
		return nil, err
	}

	if err := setColumnWidths(f, sheet, headers); err != nil {
		return nil, err
	}

	if err := addNativeChart(f, sheet, chart); err != nil {
		return nil, err
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func strPtr(s string) *string { return &s }

func writeHeader(f *excelize.File, sheet, title string, titleStyle int) (int, error) {
	row := 1
	if err := f.SetCellValue(sheet, fmt.Sprintf("A%d", row), "WMS-RSD — "+title); err != nil {
		return 0, err
	}
	if err := f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), titleStyle); err != nil {
		return 0, err
	}
	row++
	if err := f.SetCellValue(sheet, fmt.Sprintf("A%d", row), "Dicetak: "+time.Now().Format("2 January 2006 15:04")+" WIB"); err != nil {
		return 0, err
	}
	return row + 2, nil
}

type summaryRowDef struct {
	label   string
	kind    aggregateKind
	formula func(dataStartRow, dataEndRow int) string
}

func buildSummaryRows(headers []string) []summaryRowDef {
	out := []summaryRowDef{
		{
			label: "Total Baris",
			kind:  aggregateNone,
			formula: func(start, end int) string {
				return fmt.Sprintf("=COUNTA(A%d:A%d)", start, end)
			},
		},
	}
	for colIdx, header := range headers {
		kind := classifyColumn(header)
		colLetter, err := excelize.ColumnNumberToName(colIdx + 1)
		if err != nil {
			continue
		}
		switch kind {
		case aggregateCurrency:
			out = append(out, summaryRowDef{
				label: "Total " + header,
				kind:  aggregateCurrency,
				formula: func(start, end int) string {
					return fmt.Sprintf("=SUM(%s%d:%s%d)", colLetter, start, colLetter, end)
				},
			})
		case aggregateQty:
			out = append(out, summaryRowDef{
				label: "Total " + header,
				kind:  aggregateQty,
				formula: func(start, end int) string {
					return fmt.Sprintf("=SUM(%s%d:%s%d)", colLetter, start, colLetter, end)
				},
			})
		case aggregateGudang:

			out = append(out, summaryRowDef{
				label: "Gudang Terlibat",
				kind:  aggregateNone,
				formula: func(start, end int) string {
					rng := fmt.Sprintf("%s%d:%s%d", colLetter, start, colLetter, end)
					return fmt.Sprintf(`=SUMPRODUCT((%s<>"")/COUNTIF(%s,%s&""))`, rng, rng, rng)
				},
			})
		}
	}
	return out
}

type summaryStyles struct {
	bold         int
	boldCurrency int
	boldQty      int
}

func writeSummaryFormulas(
	f *excelize.File, sheet string, defs []summaryRowDef, startRow int,
	styles summaryStyles, dataStartRow, dataEndRow int,
) error {
	row := startRow
	if len(defs) == 0 {
		return nil
	}

	if err := f.SetCellValue(sheet, fmt.Sprintf("A%d", row), "Ringkasan"); err != nil {
		return err
	}
	if err := f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), styles.bold); err != nil {
		return err
	}
	row++

	for _, def := range defs {
		labelCell := fmt.Sprintf("A%d", row)
		valueCell := fmt.Sprintf("B%d", row)
		if err := f.SetCellValue(sheet, labelCell, def.label); err != nil {
			return err
		}
		if err := f.SetCellFormula(sheet, valueCell, def.formula(dataStartRow, dataEndRow)); err != nil {
			return err
		}
		style := styles.bold
		switch def.kind {
		case aggregateCurrency:
			style = styles.boldCurrency
		case aggregateQty:
			style = styles.boldQty
		}
		if err := f.SetCellStyle(sheet, valueCell, valueCell, style); err != nil {
			return err
		}
		row++
	}

	return nil
}

func writeHeaders(f *excelize.File, sheet string, headers []string, headerRow int, boldStyle int) error {
	for i, h := range headers {
		cell, err := excelize.CoordinatesToCellName(i+1, headerRow)
		if err != nil {
			return err
		}
		if err := f.SetCellValue(sheet, cell, h); err != nil {
			return err
		}
		if err := f.SetCellStyle(sheet, cell, cell, boldStyle); err != nil {
			return err
		}
	}
	return nil
}

type rowCellStyles struct {
	currency int
	qty      int
}

func writeRows(f *excelize.File, sheet string, rows [][]string, headerRow int, colKinds []aggregateKind, styles rowCellStyles) error {
	for r, dataRow := range rows {
		for c, val := range dataRow {
			if err := writeRowCell(f, sheet, c+1, headerRow+1+r, val, rowColumnKind(colKinds, c), styles); err != nil {
				return err
			}
		}
	}
	return nil
}

func rowColumnKind(colKinds []aggregateKind, column int) aggregateKind {
	if column >= 0 && column < len(colKinds) {
		return colKinds[column]
	}
	return aggregateNone
}

func writeRowCell(f *excelize.File, sheet string, column, row int, value string, kind aggregateKind, styles rowCellStyles) error {
	cell, err := excelize.CoordinatesToCellName(column, row)
	if err != nil {
		return err
	}
	if kind != aggregateCurrency && kind != aggregateQty {
		return f.SetCellValue(sheet, cell, value)
	}
	n, ok := parseNumericCell(value)
	if !ok {
		return f.SetCellValue(sheet, cell, value)
	}
	if err := f.SetCellValue(sheet, cell, n); err != nil {
		return err
	}
	style := styles.qty
	if kind == aggregateCurrency {
		style = styles.currency
	}
	return f.SetCellStyle(sheet, cell, cell, style)
}

func setColumnWidths(f *excelize.File, sheet string, headers []string) error {
	for i := range headers {
		col, err := excelize.ColumnNumberToName(i + 1)
		if err != nil {
			continue
		}
		_ = f.SetColWidth(sheet, col, col, 22)
	}
	return nil
}

func addNativeChart(f *excelize.File, sheet string, chart *ChartData) error {
	if chart == nil || len(chart.Values) == 0 {
		return nil
	}
	const helperCol = "AA"
	helperColLabel := "AB"

	if err := f.SetCellValue(sheet, helperCol+"1", "(Data Grafik — "+chart.Title+")"); err != nil {
		return err
	}
	for i, label := range chart.Labels {
		r := i + 2
		if err := f.SetCellValue(sheet, fmt.Sprintf("%s%d", helperColLabel, r), label); err != nil {
			return err
		}
		if err := f.SetCellValue(sheet, fmt.Sprintf("%s%d", helperCol, r), chart.Values[i]); err != nil {
			return err
		}
	}

	lastRow := len(chart.Values) + 1
	chartType := excelize.Col
	if chart.Type == "line" {
		chartType = excelize.Line
	}
	return f.AddChart(sheet, "AD2", &excelize.Chart{
		Type: chartType,
		Series: []excelize.ChartSeries{
			{
				Name:       sheet + "!" + helperCol + "1",
				Categories: sheet + "!" + helperColLabel + "2:" + helperColLabel + fmt.Sprint(lastRow),
				Values:     sheet + "!" + helperCol + "2:" + helperCol + fmt.Sprint(lastRow),
			},
		},
		Title: excelize.ChartTitle{
			Paragraph: []excelize.RichTextRun{{Text: chart.Title}},
		},
		Legend: excelize.ChartLegend{
			Position: "bottom",
		},
		PlotArea: excelize.ChartPlotArea{ShowVal: true},
	})
}
