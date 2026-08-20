package model

import (
	"encoding/json"
	"fmt"
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

	// FotoData/FotoContentType: foto bukti kondisi fisik barang rusak
	// disimpan LANGSUNG di database (bytea), BUKAN sebagai file statis di
	// folder storage/ yang disajikan lewat app.Static("/uploads", ...) —
	// pola SAMA PERSIS dengan alasan model.User.AvatarData: foto bukti
	// rusak bisa memuat info sensitif (kondisi gudang, lokasi, dst) dan
	// app.Static("/uploads", ...) TIDAK punya proteksi login sama sekali,
	// jadi sebelumnya siapa pun dengan URL-nya bisa lihat tanpa login.
	// Diisi lewat POST /barang-rusak/:id/foto, dibaca lewat
	// GET /barang-rusak/:id/foto (WAJIB login — lihat
	// barang_rusak_controller.go UploadFoto/ServeFoto). json:"-" karena
	// binary besar tidak boleh ikut ke-serialize ke payload JSON biasa —
	// lihat MarshalJSON di bawah, yang menghitung "foto_url" secara
	// dinamis (path endpoint, BUKAN data binary-nya) mengikuti pola
	// User.AvatarURL().
	FotoData        []byte `json:"-" gorm:"type:bytea"`
	FotoContentType string `json:"-" gorm:"size:100"`

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

// FotoURL menghitung URL untuk mengambil foto bukti lewat
// GET /barang-rusak/:id/foto (WAJIB login — lihat
// internal/controller/barang_rusak ServeFoto). Padanan langsung dari
// model.User.AvatarURL() — lihat catatan lengkap di field FotoData di
// atas kenapa fotonya disimpan di database, bukan file statis. Kosong
// berarti belum ada foto diunggah — frontend jatuh ke placeholder.
// Query string ?v=<updated_at> memaksa browser ambil ulang gambar
// setelah foto diganti (nama URL-nya sendiri tidak berubah).
func (b BarangRusak) FotoURL() string {
	if len(b.FotoData) == 0 {
		return ""
	}
	return fmt.Sprintf("/barang-rusak/%d/foto?v=%d", b.ID, b.UpdatedAt.Unix())
}

// MarshalJSON menyisipkan field "foto_url" HASIL HITUNGAN (dari method
// FotoURL() di atas) ke output JSON, tanpa perlu Response DTO terpisah di
// controller (handler-handler modul ini masih mengembalikan *BarangRusak
// langsung, lihat barang_rusak_controller.go). Trik type-alias supaya
// TIDAK rekursi tak berhingga (memanggil json.Marshal(b) lagi di dalam
// method MarshalJSON milik BarangRusak sendiri).
func (b BarangRusak) MarshalJSON() ([]byte, error) {
	type Alias BarangRusak
	return json.Marshal(struct {
		Alias
		FotoURL string `json:"foto_url"`
	}{
		Alias:   Alias(b),
		FotoURL: b.FotoURL(),
	})
}
