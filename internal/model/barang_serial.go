package model

import (
	"time"

	"gorm.io/gorm"
)

// BarangSerial — satu unit FISIK barang yang dilacak lewat nomor seri
// (SN), supaya dua barang dengan KodeBarang/SKU yang SAMA tapi unit
// fisiknya berbeda tetap bisa dibedakan — kebutuhan umum gudang ISP,
// mis. dua ONT model yang sama persis di rak yang sama, tapi SN/MAC
// address perangkatnya beda dan salah satunya sudah pernah dipakai di
// lokasi pelanggan tertentu.
//
// HANYA barang dengan Barang.IsSerialized = true yang punya baris di
// tabel ini. Barang non-serial (kabel per meter, baut, ATK, dst) tetap
// cukup pakai Barang.Stok agregat seperti biasa — modul ini SENGAJA
// tidak dipaksakan ke semua barang supaya operator tidak direpotkan
// input SN untuk barang yang memang tidak butuh (lihat toggle
// IsSerialized di form Tambah/Ubah Barang).
//
// Siklus hidup satu baris:
//  1. Dibuat saat dokumen Barang Masuk diselesaikan (Complete) — operator
//     input/scan SN fisik sejumlah Qty pada item yang barangnya
//     IsSerialized. Status = "tersedia", GudangID = gudang tujuan.
//  2. Saat Barang Keluar diselesaikan, operator MEMILIH SN spesifik
//     (bukan cuma qty) yang keluar dari gudang tersebut. Status berubah
//     jadi "terpasang", GudangID/RakID dikosongkan (bukan lagi tercatat
//     "di gudang mana", karena sudah keluar ke lapangan/pelanggan).
//  3. Bisa juga ditandai "rusak" manual (lihat
//     internal/controller/barang_serial UpdateStatus) tanpa harus lewat
//     dokumen Barang Keluar — mis. ditemukan cacat produksi saat stock
//     opname.
//
// Baris di sini TIDAK mengubah cara Barang.Stok dihitung (tetap agregat,
// diupdate di tempat yang sama seperti sebelumnya) — tabel ini murni
// lapisan rincian PER UNIT di atasnya, supaya satu SN tertentu bisa
// dicari/dilacak riwayatnya tanpa mengubah alur stok yang sudah ada.
type BarangSerial struct {
	ID       uint    `json:"id" gorm:"primaryKey"`
	BarangID uint    `json:"barang_id" gorm:"not null;index"`
	Barang   *Barang `json:"barang,omitempty" gorm:"foreignKey:BarangID"`

	// SerialNumber: nomor seri fisik (label produsen, MAC address, dsb).
	// UNIK secara global (bukan hanya per-barang) — satu SN fisik hanya
	// boleh terdaftar sekali di seluruh sistem, walau nantinya bisa
	// pindah status (tersedia -> terpasang -> rusak) atau pindah gudang.
	SerialNumber string `json:"serial_number" gorm:"size:100;not null;uniqueIndex"`

	// Status: tersedia (ada fisik di GudangID/RakID, siap dikeluarkan) |
	// terpasang (sudah keluar lewat Barang Keluar, terpasang di
	// lapangan/pelanggan) | rusak (ditandai manual, tidak bisa
	// dikeluarkan lagi selama status ini).
	Status string `json:"status" gorm:"size:20;not null;default:'tersedia';index"`

	// GudangID/RakID: lokasi fisik SAAT INI. Nil kalau unit sudah
	// "terpasang" (keluar gudang, tidak lagi tercatat ada di rak
	// manapun) — beda dengan model.Asset yang punya koordinat GPS
	// permanen; unit barang di sini cuma dilacak "ada di gudang mana"
	// sebelum keluar, bukan titik lokasi presisi di lapangan.
	GudangID *uint   `json:"gudang_id" gorm:"index"`
	Gudang   *Gudang `json:"gudang,omitempty" gorm:"foreignKey:GudangID"`
	RakID    *uint   `json:"rak_id" gorm:"index"`
	Rak      *Rak    `json:"rak,omitempty" gorm:"foreignKey:RakID"`

	// BarangMasukItemID/BarangKeluarItemID: jejak dokumen asal unit ini
	// masuk & dokumen terakhir yang mengeluarkannya — dasar telusur
	// riwayat satu SN tanpa perlu tabel histori terpisah (lihat
	// internal/controller/barang_serial Riwayat(), yang menggabungkan
	// info dari kedua dokumen ini).
	BarangMasukItemID  *uint             `json:"barang_masuk_item_id" gorm:"index"`
	BarangMasukItem    *BarangMasukItem  `json:"barang_masuk_item,omitempty" gorm:"foreignKey:BarangMasukItemID"`
	BarangKeluarItemID *uint             `json:"barang_keluar_item_id" gorm:"index"`
	BarangKeluarItem   *BarangKeluarItem `json:"barang_keluar_item,omitempty" gorm:"foreignKey:BarangKeluarItemID"`

	Catatan string `json:"catatan" gorm:"size:255"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	// DeletedAt: soft-delete GORM — lihat catatan lengkap di
	// model.BarangRusak. SN yang dihapus (mis. salah input) tetap bisa
	// dipulihkan lewat Tempat Sampah, bukan hilang permanen begitu saja.
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (BarangSerial) TableName() string { return "barang_serials" }
