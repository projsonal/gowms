package constant

// Status dokumen Stock Opname (penghitungan fisik stok).
const (
	StatusSODraft      = "draft"
	StatusSOSelesai    = "selesai"
	StatusSODibatalkan = "dibatalkan"
)

const (
	ErrSOTidakDitemukan = "dokumen stock opname tidak ditemukan"
	ErrSOBukanDraft     = "dokumen stock opname masih berupa draft"
)
