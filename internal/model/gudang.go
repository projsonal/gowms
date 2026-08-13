package model

import "time"

type Kategori struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Nama      string    `json:"nama" gorm:"size:100;uniqueIndex;not null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Kategori) TableName() string { return "kategori" }

type Satuan struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Nama      string    `json:"nama" gorm:"size:50;uniqueIndex;not null"`
	Singkatan string    `json:"singkatan" gorm:"size:10;not null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Satuan) TableName() string { return "satuan" }

type Gudang struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Nama        string    `json:"nama" gorm:"size:100;not null"`
	// Kode: kode singkat gudang (mis. "BBU", "MAHANG") — dipakai sebagai
	// prefix label RSD aset gudang, format {KODE}-RSD-{nomor urut per
	// gudang}, lihat model.Asset & internal/controller/asset. Disimpan
	// UPPERCASE, unik per gudang (lihat gudang_controller.go Create/Update).
	Kode        string    `json:"kode" gorm:"size:20;uniqueIndex"`
	Alamat      string    `json:"alamat" gorm:"size:255"`
	// PIC & Kapasitas: SEBELUMNYA tidak ada di model sama sekali walau
	// tabel Manajemen Gudang di frontend SUDAH LAMA menampilkan kolom
	// "PIC" dan "KAPASITAS" — keduanya cuma placeholder statis ("-" dan
	// "0/0") karena memang tidak ada tempat menyimpannya. Kapasitas di
	// sini SENGAJA "kapasitas total" yang diisi manual (bukan otomatis
	// dihitung dari stok barang) — model Barang tidak punya relasi
	// gudang_id (stok agregat lintas gudang, lihat catatan panjang di
	// lib/api/mappers.ts frontend), jadi "kapasitas TERPAKAI" per gudang
	// belum bisa dihitung otomatis tanpa perubahan skema lebih besar
	// (tabel barang_gudang) — itu di luar cakupan perbaikan kali ini.
	PIC         string    `json:"pic" gorm:"size:100"`
	// Telepon: nomor kontak gudang — dipakai sebagai "No. Telepon Pengirim"
	// di resi pengiriman (Receipt.tsx) saat pengiriman berasal dari gudang
	// ini. Opsional karena gudang lama belum tentu sudah diisi.
	Telepon     string    `json:"telepon" gorm:"size:20"`
	Kapasitas   int       `json:"kapasitas" gorm:"not null;default:0"`
	// Latitude/Longitude: opsional — dipakai untuk menampilkan titik lokasi
	// gudang di peta (lihat DeliveriesMap.tsx frontend, dipakai berdampingan
	// dengan marker posisi kurir). Nullable karena gudang lama belum tentu
	// sudah diisi koordinatnya lewat form Tambah/Ubah Gudang.
	Latitude    *float64  `json:"latitude"`
	Longitude   *float64  `json:"longitude"`
	IsProtected bool      `json:"is_protected" gorm:"not null;default:false"` // dikunci super_admin — lihat internal/controller/gudang Protect()
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Raks        []Rak     `json:"raks,omitempty" gorm:"foreignKey:GudangID"`
}

func (Gudang) TableName() string { return "gudangs" }

type Rak struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	KodeRak   string    `json:"kode_rak" gorm:"size:20;uniqueIndex;not null"` // A2-R05
	GudangID  uint      `json:"gudang_id" gorm:"not null;index"`
	Gudang    *Gudang   `json:"gudang,omitempty" gorm:"foreignKey:GudangID"`
	Kapasitas int       `json:"kapasitas" gorm:"not null;default:0"`
	Terisi    int       `json:"terisi" gorm:"not null;default:0"`
	Status    string    `json:"status" gorm:"size:20;default:'kosong'"` // kosong|terisi_sebagian|penuh
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Rak) TableName() string { return "raks" }

func (r *Rak) RecalculateStatus() {
	switch {
	case r.Terisi <= 0:
		r.Status = "kosong"
	case r.Terisi >= r.Kapasitas:
		r.Status = "penuh"
	default:
		r.Status = "terisi_sebagian"
	}
}
