package model

import "time"

type Pengiriman struct {
	ID               uint          `json:"id" gorm:"primaryKey"`
	NomorPengiriman  string        `json:"nomor_pengiriman" gorm:"size:30;uniqueIndex;not null"` // PG-2026-0001
	BarangKeluarID   *uint         `json:"barang_keluar_id" gorm:"index"`
	BarangKeluar     *BarangKeluar `json:"barang_keluar,omitempty" gorm:"foreignKey:BarangKeluarID"`
	GudangAsalID     uint          `json:"gudang_asal_id" gorm:"not null;index"`
	GudangAsal       *Gudang       `json:"gudang_asal,omitempty" gorm:"foreignKey:GudangAsalID"`
	JenisPengambilan string        `json:"jenis_pengambilan" gorm:"size:10;not null"` // pickup | dropoff
	NamaPenerima     string        `json:"nama_penerima" gorm:"size:150;not null"`
	TeleponPenerima  string        `json:"telepon_penerima" gorm:"size:20"`
	AlamatTujuan     string        `json:"alamat_tujuan" gorm:"size:255"` // wajib untuk dropoff
	// Latitude/Longitude TUJUAN (opsional) — TANPA ini, peta pelacakan
	// tidak bisa menggambar rute/marker tujuan sama sekali (cuma ada
	// posisi kurir & gudang asal). Diisi manual saat Tambah/Ubah jadwal
	// (mis. disalin dari Google Maps), sama pola dengan Gudang.Latitude.
	DestLat          *float64      `json:"dest_lat"`
	DestLng          *float64      `json:"dest_lng"`
	NamaKurir        string        `json:"nama_kurir" gorm:"size:100"`
	TeleponKurir     string        `json:"telepon_kurir" gorm:"size:20"`
	Status           string        `json:"status" gorm:"size:20;not null;default:'draft';index"`
	TanggalKirim     time.Time     `json:"tanggal_kirim" gorm:"not null"`
	EstimasiTiba     *time.Time    `json:"estimasi_tiba"`
	WaktuTerkirim    *time.Time    `json:"waktu_terkirim"`
	Catatan          string        `json:"catatan" gorm:"size:255"`
	IsProtected      bool          `json:"is_protected" gorm:"not null;default:false"` // dikunci super_admin

	// Posisi kurir terakhir — diisi lewat endpoint POST /pengiriman/:id/lokasi
	// selama status "dalam_perjalanan". Murni dari input GPS device kurir
	// (aplikasi mobile/browser), tidak ada ketergantungan hardware khusus.
	LastLat        *float64   `json:"last_lat"`
	LastLng        *float64   `json:"last_lng"`
	LastLocationAt *time.Time `json:"last_location_at"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Pengiriman) TableName() string { return "pengiriman" }

func (p *Pengiriman) SudahDikirim() bool {
	return p.Status == "terkirim" || p.Status == "dibatalkan"
}

type PengirimanTrackingPoint struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	PengirimanID uint      `json:"pengiriman_id" gorm:"not null;index"`
	Lat          float64   `json:"lat" gorm:"not null"`
	Lng          float64   `json:"lng" gorm:"not null"`
	KecepatanKmh *float64  `json:"kecepatan_kmh"` // opsional, kecepatan sesaat km/jam
	RecordedAt   time.Time `json:"recorded_at" gorm:"not null;index"`
	CreatedAt    time.Time `json:"created_at"`
}

func (PengirimanTrackingPoint) TableName() string { return "pengiriman_tracking_points" }
