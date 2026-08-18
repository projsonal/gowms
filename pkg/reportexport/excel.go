package reportexport

import (
	"fmt"
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

	row, err := writeHeader(f, sheet, title, titleStyle)
	if err != nil {
		return nil, err
	}

	row, err = writeSummary(f, sheet, summary, row, boldStyle)
	if err != nil {
		return nil, err
	}

	if err := writeHeaders(f, sheet, headers, row, boldStyle); err != nil {
		return nil, err
	}

	if err := writeRows(f, sheet, headers, rows, row); err != nil {
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

func writeSummary(f *excelize.File, sheet string, summary [][2]string, startRow int, boldStyle int) (int, error) {
	row := startRow
	if len(summary) == 0 {
		return row, nil
	}

	if err := f.SetCellValue(sheet, fmt.Sprintf("A%d", row), "Ringkasan"); err != nil {
		return 0, err
	}
	if err := f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), boldStyle); err != nil {
		return 0, err
	}
	row++

	for _, kv := range summary {
		if err := f.SetCellValue(sheet, fmt.Sprintf("A%d", row), kv[0]); err != nil {
			return 0, err
		}
		if err := f.SetCellValue(sheet, fmt.Sprintf("B%d", row), kv[1]); err != nil {
			return 0, err
		}
		row++
	}

	return row + 1, nil
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

func writeRows(f *excelize.File, sheet string, _ []string, rows [][]string, headerRow int) error {
	for r, dataRow := range rows {
		for c, val := range dataRow {
			cell, err := excelize.CoordinatesToCellName(c+1, headerRow+1+r)
			if err != nil {
				return err
			}
			if err := f.SetCellValue(sheet, cell, val); err != nil {
				return err
			}
		}
	}
	return nil
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
	const helperCol = "AA" // jauh dari kolom data laporan (laporan di sini tidak pernah >15 kolom)
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
