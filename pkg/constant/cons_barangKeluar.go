package constant

const (
	StatusBKDraft      = "draft"
	StatusBKSelesai    = "selesai"
	StatusBKDibatalkan = "dibatalkan"
)

const (
	ErrBKTidakDitemukan = "dokumen barang keluar tidak ditemukan"
	ErrBKBukanDraft     = "dokumen barang keluar hanya bisa diubah/dihapus/diselesaikan selama masih berstatus draft"
	ErrBKStokTidakCukup = "stok barang tidak mencukupi untuk pengeluaran ini"
)
