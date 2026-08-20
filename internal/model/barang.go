package model

import (
	"time"

	"gorm.io/gorm"
)

type Barang struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	KodeBarang  string    `json:"kode_barang" gorm:"size:30;uniqueIndex;not null"` // SKU, mis: "BRG-0001"
	Nama        string    `json:"nama" gorm:"size:150;not null;index"`
	KategoriID  uint      `json:"kategori_id" gorm:"not null;index"`
	Kategori    *Kategori `json:"kategori,omitempty" gorm:"foreignKey:KategoriID"`
	SatuanID    uint      `json:"satuan_id" gorm:"not null;index"`
	Satuan      *Satuan   `json:"satuan,omitempty" gorm:"foreignKey:SatuanID"`
	HargaBeli   int64     `json:"harga_beli" gorm:"not null;default:0"`   // rupiah/unit
	StokMinimum int       `json:"stok_minimum" gorm:"not null;default:0"` // reorder point; 0 = tidak dipantau
	Stok        int       `json:"stok" gorm:"not null;default:0"`         // agregat total, lihat AdjustStok()
	// BeratGram: berat satuan barang dalam gram (opsional) — dipakai
	// menampilkan "Berat" per item di resi pengiriman (Receipt.tsx), mirip
	// label J&T/Shopee. Nil kalau belum diisi lewat form Tambah/Ubah Barang.
	BeratGram   *int   `json:"berat_gram"`
	IsActive    bool   `json:"is_active" gorm:"not null;default:true"`     // nonaktif = didiskontinu, tetap tampil di histori
	IsProtected bool   `json:"is_protected" gorm:"not null;default:false"` // dikunci super_admin — lihat internal/controller/barang Protect()
	Deskripsi   string `json:"deskripsi" gorm:"size:255"`

	// IsSerialized: true kalau unit fisik barang ini WAJIB dibedakan lewat
	// nomor seri (SN) — umum di gudang ISP: dua ONT/modem dengan
	// KodeBarang (SKU) yang SAMA tapi unit fisiknya beda (mis. MAC
	// address / SN produsen beda-beda), jadi tidak bisa dianggap
	// "identik" begitu saja seperti stok barang biasa (kabel per meter,
	// baut, dst). Kalau true: Barang Masuk/Keluar untuk barang ini WAJIB
	// menyertakan daftar SN sejumlah Qty saat diselesaikan (lihat
	// model.BarangSerial & internal/controller/barang_masuk|barang_keluar
	// Complete()). Stok agregat (field Stok di atas) TETAP dipakai
	// sebagai angka ringkas, BarangSerial adalah rincian per-unitnya —
	// keduanya harus selalu sinkron (dijaga lewat validasi di Complete()).
	IsSerialized bool `json:"is_serialized" gorm:"not null;default:false"`

	// --- Alur persetujuan (khusus barang yang DIBUAT role admin) ---
	// super_admin membuat barang -> langsung "disetujui" (default).
	// admin membuat barang       -> otomatis "menunggu" sampai super_admin
	//                                Approve/Reject, lihat Create()/Approve()/
	//                                Reject() di barang_controller.go.
	// karyawan HANYA melihat barang berstatus "disetujui" (lihat List()).
	ApprovalStatus  string     `json:"approval_status" gorm:"size:20;not null;default:'disetujui';index"` // disetujui | menunggu | ditolak
	DiajukanOleh    *uint      `json:"diajukan_oleh"`
	DisetujuiOleh   *uint      `json:"disetujui_oleh"`
	CatatanApproval string     `json:"catatan_approval" gorm:"size:255"`
	DireviewPada    *time.Time `json:"direview_pada"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	// DeletedAt: soft-delete GORM — lihat catatan lengkap di model.BarangRusak.
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (Barang) TableName() string { return "barang" }

func (b *Barang) IsStokMenipis() bool {
	return b.StokMinimum > 0 && b.Stok <= b.StokMinimum
}

func (b *Barang) NilaiInventaris() int64 {
	return int64(b.Stok) * b.HargaBeli
}
