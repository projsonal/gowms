package reportexport

import (
	"fmt"
	"time"

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

func ToExcel(title string, summary [][2]string, headers []string, rows [][]string) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	sheet := sanitizeSheetName(title)
	if err := f.SetSheetName(f.GetSheetName(0), sheet); err != nil {
		return nil, err
	}

	// Catatan: pengaturan page-layout (ukuran kertas/orientasi cetak)
	// SENGAJA tidak diset lewat API di sini — versi excelize yang dipakai
	// belum bisa dipastikan cocok tanpa akses toolchain Go di lingkungan
	// pengembangan ini (lihat catatan verifikasi build di ringkasan).
	// Excel/LibreOffice tetap bisa mengatur cetak A4 potrait manual dari
	// dialog Page Setup — konten (kop surat WMS-RSD & ringkasan) yang
	// lebih penting sudah tetap ada di baris pertama sheet ini.

	boldStyle, err := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})
	if err != nil {
		return nil, err
	}
	titleStyle, err := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true, Size: 14, Color: "146C14"}})
	if err != nil {
		return nil, err
	}

	row := 1
	if err := f.SetCellValue(sheet, fmt.Sprintf("A%d", row), "WMS-RSD — "+title); err != nil {
		return nil, err
	}
	if err := f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), titleStyle); err != nil {
		return nil, err
	}
	row++
	if err := f.SetCellValue(sheet, fmt.Sprintf("A%d", row), "Dicetak: "+time.Now().Format("2 January 2006 15:04")+" WIB"); err != nil {
		return nil, err
	}
	row += 2

	if len(summary) > 0 {
		if err := f.SetCellValue(sheet, fmt.Sprintf("A%d", row), "Ringkasan"); err != nil {
			return nil, err
		}
		if err := f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), boldStyle); err != nil {
			return nil, err
		}
		row++
		for _, kv := range summary {
			if err := f.SetCellValue(sheet, fmt.Sprintf("A%d", row), kv[0]); err != nil {
				return nil, err
			}
			if err := f.SetCellValue(sheet, fmt.Sprintf("B%d", row), kv[1]); err != nil {
				return nil, err
			}
			row++
		}
		row++
	}

	headerRow := row
	for i, h := range headers {
		cell, err := excelize.CoordinatesToCellName(i+1, headerRow)
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

	for r, dataRow := range rows {
		for c, val := range dataRow {
			cell, err := excelize.CoordinatesToCellName(c+1, headerRow+1+r)
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
