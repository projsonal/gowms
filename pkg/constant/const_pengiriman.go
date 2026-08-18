package constant

const (
	StatusPGDraft           = "Draft"
	StatusPGDijadwalkan     = "Dijadwalkan"
	StatusPGDalamPerjalanan = "Dalam Perjalanan"
	StatusPGTerkirim        = "Terkirim"
	StatusPGSelesai         = "Selesai"
	StatusPGDibatalkan      = "Dibatalkan"
)

const (
	JenisPickup  = "Pickup"
	JenisDropoff = "Dropoff"
)

const (
	ErrPGTidakDitemukan   = "Dokumen pengiriman tidak ditemukan."
	ErrPGBukanDraft       = "Dokumen pengiriman sudah diproses lebih lanjut — aksi ini hanya bisa dilakukan saat statusnya masih Draft."
	ErrPGBukanDijadwalkan = "Pengiriman perlu dijadwalkan terlebih dahulu sebelum barang di kirimkan."
	ErrPGBukanPerjalan    = "Barang masih dalam tahap packing."
	ErrPGAlamatWajib      = "Alamat tujuan wajib di isi untuk pengiriman jenid dropoff"
)
