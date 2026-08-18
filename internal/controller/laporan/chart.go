package laporan

import (
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/projsonal/gowms/pkg/constant"
)

// ChartData — hasil agregasi generik untuk "Analisa Data" tiap laporan:
// dipakai buat chart di UI (recharts, frontend) DAN diselipkan ke file
// unduhan (PDF/Excel/Word) supaya angkanya konsisten di semua tempat.
type ChartData struct {
	Title  string    `json:"title"`
	Type   string    `json:"type"` // "bar" | "line"
	Labels []string  `json:"labels"`
	Values []float64 `json:"values"`
}

// Insight — ringkasan teks otomatis dari ChartData, dipakai sebagai
// FALLBACK saat chart-nya sendiri tidak bisa disisipkan ke format file
// tertentu (khususnya .docx — lihat pkg/reportexport/docs.go, dirakit
// manual dari OOXML mentah tanpa library chart) — daripada tidak ada
// analisa sama sekali di file itu, tetap ada ringkasan tekstual yang
// mengandung informasi yang SAMA dengan yang divisualisasikan chart.
func (cd *ChartData) Insight() string {
	if cd == nil || len(cd.Values) == 0 {
		return "Belum ada data yang cukup untuk dianalisis pada periode ini."
	}
	total := 0.0
	maxIdx, minIdx := 0, 0
	for i, v := range cd.Values {
		total += v
		if v > cd.Values[maxIdx] {
			maxIdx = i
		}
		if v < cd.Values[minIdx] {
			minIdx = i
		}
	}
	avg := total / float64(len(cd.Values))
	return "Total keseluruhan: " + trimFloat(total) + ". Rata-rata per periode: " + trimFloat(avg) +
		". Tertinggi: " + cd.Labels[maxIdx] + " (" + trimFloat(cd.Values[maxIdx]) + "). Terendah: " +
		cd.Labels[minIdx] + " (" + trimFloat(cd.Values[minIdx]) + ")."
}

func trimFloat(f float64) string {
	s := strconv.FormatFloat(f, 'f', 1, 64)
	s = strings.TrimSuffix(s, ".0")
	return s
}

// Granularity yang didukung untuk chart deret waktu (Harian/Bulanan/Tahunan).
const (
	GranularitasHarian  = "harian"
	GranularitasBulanan = "bulanan"
	GranularitasTahunan = "tahunan"
)

func normalizeGranularity(g string) string {
	switch g {
	case GranularitasHarian, GranularitasTahunan:
		return g
	default:
		return GranularitasBulanan
	}
}

func periodKey(t time.Time, granularity string) (key, label string) {
	switch granularity {
	case GranularitasHarian:
		return t.Format("2006-01-02"), t.Format("2 Jan 2006")
	case GranularitasTahunan:
		return t.Format("2006"), t.Format("2006")
	default: // bulanan
		return t.Format("2006-01"), t.Format("Jan 2006")
	}
}

// computeDateSeriesChart menghitung JUMLAH BARIS (transaksi/dokumen) per
// periode (hari/bulan/tahun) dari kolom tanggal — dipakai generik untuk
// SEMUA laporan berbasis tanggal (Barang Keluar, Barang Masuk, Barang
// Retur, Purchase Order, Stock Opname), supaya "Analisa Data" tersedia di
// tiap menu Laporan tanpa logika terpisah per modul. `dateColCandidates`
// berisi kemungkinan nama header kolom tanggal (dicoba berurutan, dipakai
// yang pertama cocok) karena tiap laporan menamai kolom tanggalnya beda.
func computeDateSeriesChart(title string, headers []string, rows [][]string, dateColCandidates []string, granularity string) *ChartData {
	granularity = normalizeGranularity(granularity)
	dateColIdx := -1
	for _, candidate := range dateColCandidates {
		for i, h := range headers {
			if h == candidate {
				dateColIdx = i
				break
			}
		}
		if dateColIdx >= 0 {
			break
		}
	}
	if dateColIdx < 0 {
		return nil
	}

	counts := map[string]float64{}
	labels := map[string]string{}
	for _, row := range rows {
		if dateColIdx >= len(row) {
			continue
		}
		t, err := time.Parse(dateFormat, row[dateColIdx])
		if err != nil {
			// beberapa kolom tanggal diformat "2 Jan 2006" bukan
			// dateFormat mentah — coba format alternatif sebelum menyerah
			// pada baris ini.
			t, err = time.Parse("2 January 2006", row[dateColIdx])
			if err != nil {
				continue
			}
		}
		key, label := periodKey(t, granularity)
		counts[key]++
		labels[key] = label
	}
	if len(counts) == 0 {
		return nil
	}

	keys := make([]string, 0, len(counts))
	for k := range counts {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	cd := &ChartData{Title: title, Type: "bar"}
	for _, k := range keys {
		cd.Labels = append(cd.Labels, labels[k])
		cd.Values = append(cd.Values, counts[k])
	}
	return cd
}

// computeTopStokChart — khusus Laporan Stok Barang (bukan deret waktu,
// snapshot kondisi stok saat ini): top 10 barang dengan stok TERBANYAK,
// bar chart menurun. Mengasumsikan kolom "Stok" ada di headers (lihat
// buildStokBarang).
func computeTopStokChart(headers []string, rows [][]string) *ChartData {
	nameIdx, stokIdx := -1, -1
	for i, h := range headers {
		switch h {
		case "Nama":
			nameIdx = i
		case "Stok":
			stokIdx = i
		}
	}
	if nameIdx < 0 || stokIdx < 0 {
		return nil
	}

	type pair struct {
		name string
		stok float64
	}
	pairs := make([]pair, 0, len(rows))
	for _, row := range rows {
		if nameIdx >= len(row) || stokIdx >= len(row) {
			continue
		}
		stok, err := strconv.ParseFloat(strings.TrimSpace(row[stokIdx]), 64)
		if err != nil {
			continue
		}
		pairs = append(pairs, pair{name: row[nameIdx], stok: stok})
	}
	sort.Slice(pairs, func(i, j int) bool { return pairs[i].stok > pairs[j].stok })
	if len(pairs) > 10 {
		pairs = pairs[:10]
	}
	if len(pairs) == 0 {
		return nil
	}

	cd := &ChartData{Title: "Top 10 Barang dengan Stok Terbanyak", Type: "bar"}
	for _, p := range pairs {
		cd.Labels = append(cd.Labels, p.name)
		cd.Values = append(cd.Values, p.stok)
	}
	return cd
}

// buildChart — dispatcher analog buildReport, memilih agregator yang
// sesuai untuk tiap tipe laporan ("sesuaikan dari setiap menunya").
func (h *Controller) buildChart(tipe string, headers []string, rows [][]string, granularity string) *ChartData {
	switch tipe {
	case constant.LaporanBarangKeluar:
		return computeDateSeriesChart("Barang Keluar per Periode", headers, rows, []string{"Tanggal"}, granularity)
	case constant.LaporanBarangMasuk:
		return computeDateSeriesChart("Barang Masuk per Periode", headers, rows, []string{"Tanggal"}, granularity)
	case constant.LaporanBarangRetur:
		return computeDateSeriesChart("Barang Retur per Periode", headers, rows, []string{"Tanggal Diperiksa"}, granularity)
	case constant.LaporanPO:
		return computeDateSeriesChart("Purchase Order per Periode", headers, rows, []string{"Tanggal PO"}, granularity)
	case constant.LaporanStokOpname:
		return computeDateSeriesChart("Stock Opname per Periode", headers, rows, []string{"Tanggal"}, granularity)
	case constant.LaporanStokBarang:
		return computeTopStokChart(headers, rows)
	default:
		return nil
	}
}
