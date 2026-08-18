package constant

const (
	LaporanStokBarang   = "Stok Barang"
	LaporanBarangMasuk  = "Barng Masuk"
	LaporanBarangKeluar = "Barang Keluar"
	LaporanPO           = "Purchase Order"
	LaporanStokOpname   = "Stock Opname"
	// LaporanBarangRetur: sumbernya model.BarangRusak dengan Status="retur"
	// — barang yang dilaporkan rusak TAPI setelah pengecekan fisik masih
	// bisa diperbaiki/diretur ke supplier (lihat dokumentasi alur di
	// model.BarangRusak). Sebelum ini laporan belum punya sumber data sama
	// sekali di backend.
	LaporanBarangRetur = "Barang Retur"
)

const (
	FormatExcel = "Excel"
	FormatPDF   = "PDF"
	FormatWord  = "Docs"
)

const (
	ErrLaporanTipeTidakDidukung   = "Jenis Laporan tidak di dukung."
	ErrLaporanFormatTidakDidukung = "Format ekspor laporan tidak di dukung (Gunakan excel, pdf, atau docs)."
)
