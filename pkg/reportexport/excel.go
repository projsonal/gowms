package reportexport

import (
	"github.com/xuri/excelize/v2"
)

const maxSheetNameLen = 31 // batas panjang nama sheet Excel

func sanitizeSheetName(title string) string {
	if title == "" {
		return "Laporan"
	}
	if len(title) > maxSheetNameLen {
		return title[:maxSheetNameLen]
	}
	return title
}

func ToExcel(title string, headers []string, rows [][]string) ([]byte, error) {
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

	for i, h := range headers {
		cell, err := excelize.CoordinatesToCellName(i+1, 1)
		if err != nil {
			return nil, err
		}
		if err := f.SetCellValue(sheet, cell, h); err != nil {
			return nil, err
		}
		if err := f.SetCellStyle(sheet, cell, cell, boldStyle); err != nil {
			return nil, err
		}
	}

	for r, row := range rows {
		for c, val := range row {
			cell, err := excelize.CoordinatesToCellName(c+1, r+2)
			if err != nil {
				return nil, err
			}
			if err := f.SetCellValue(sheet, cell, val); err != nil {
				return nil, err
			}
		}
	}

	for i := range headers {
		col, err := excelize.ColumnNumberToName(i + 1)
		if err != nil {
			continue
		}
		_ = f.SetColWidth(sheet, col, col, 22)
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
