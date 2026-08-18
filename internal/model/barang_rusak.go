package model

import (
	"time"

	"gorm.io/gorm"
)

// BarangRusak — modul Barang Rusak: pencatatan barang (SKU dari modul
// Kelola Barang) atau aset gudang yang dilaporkan rusak, sebelum
// diputuskan apakah bisa diretur ke supplier atau harus di-write-off.
//
// Alur status:
//  1. Dibuat -> Status default "pengecekan" (menunggu pemeriksaan fisik).
//  2. Setelah dicek fisik, petugas mengisi JenisBarang:
//     - "retur" -> barang MASIH BISA diretur ke supplier -> Status ikut
//     menjadi "retur".
//     - "rusak" -> barang TIDAK BISA diretur (rusak total) -> Status ikut
//     menjadi "rusak".
type BarangRusak struct {
	ID uint `json:"id" gorm:"primaryKey"`

	// BarangID: relasi opsional ke SKU di modul Kelola Barang (model.Barang).
	// Nil kalau yang dilaporkan rusak adalah aset gudang (model.Asset) atau
	// barang yang belum/tidak terdaftar sebagai SKU.
	BarangID *uint   `json:"barang_id"`
	Barang   *Barang `json:"barang,omitempty" gorm:"foreignKey:BarangID"`

	// LabelBarang: label/kode identitas barang yang dilaporkan rusak —
	// bisa kode_barang (SKU), label_rsd aset, atau kode_ba aset transportasi.
	LabelBarang string `json:"label_barang" gorm:"size:60;not null;index"`
	NamaBarang  string `json:"nama_barang" gorm:"size:150;not null"`
	Keterangan  string `json:"keterangan" gorm:"size:500"`

	// FotoURL: path relatif ke foto bukti kondisi fisik barang rusak
	// (lihat POST /barang-rusak/:id/foto, pola sama seperti avatar user —
	// disimpan di StorageConfig.Path/barang-rusak/, disajikan statis lewat
	// app.Static("/uploads", ...)). Opsional — kosong kalau belum diunggah.
	FotoURL string `json:"foto_url" gorm:"size:255"`

	// JenisBarang: hasil klasifikasi SETELAH pengecekan fisik —
	// "retur" (bisa diretur ke supplier) | "rusak" (tidak bisa diretur).
	// Kosong selama Status masih "pengecekan".
	JenisBarang string `json:"jenis_barang" gorm:"size:10"`

	// Status: pengecekan (default) | retur | rusak — lihat dokumentasi
	// alur di atas. Selalu sinkron dengan JenisBarang setelah diperiksa.
	Status string `json:"status" gorm:"size:20;not null;default:'pengecekan';index"`

	DilaporkanOleh uint       `json:"dilaporkan_oleh" gorm:"not null"`
	Pelapor        *User      `json:"pelapor,omitempty" gorm:"foreignKey:DilaporkanOleh"`
	DicekOleh      *uint      `json:"dicek_oleh"`
	Pemeriksa      *User      `json:"pemeriksa,omitempty" gorm:"foreignKey:DicekOleh"`
	DicekPada      *time.Time `json:"dicek_pada"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	// DeletedAt: soft-delete GORM — "hapus" dari UI cuma menandai baris ini
	// (bukan DELETE SQL sungguhan), supaya bisa dipulihkan lewat fitur
	// Tempat Sampah (lihat internal/controller/trash). Query normal
	// (List/FindByID/dst) otomatis mengecualikan baris yang sudah
	// soft-deleted — GORM menambahkan `WHERE deleted_at IS NULL` sendiri.
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (BarangRusak) TableName() string { return "barang_rusak" }
