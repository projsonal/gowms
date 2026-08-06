package constant

const (
	StatusPODraft      = "Draft"
	StatusPODiajukan   = "Diajukan"
	StatusPODisetujui  = "Disetujui"
	StatusPODitolak    = "Ditolak"
	StatusPOSelesai    = "Delesai"
	StatusPODibatalkan = "Dibatalkan"
)

const (
	ErrPOTidakDitemukan = "PO(Purchaese Order) tidak ditemukan"
	ErrPOBukanDraft     = "PO(Purchase Order) masih berupa draft"
	ErrPOTidakDiajukan  = "PO(Purchase Order) perlu diajukan terlebih dahulu hingga berstatus disetujui ataupun ditolak"
	ErrPOTidakDisetujui = "PO(Purchase Order) perlu disetujui terlebih dahulu sebelum barang dapat diterima"
)
