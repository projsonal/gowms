package model

import "time"

type Supplier struct {
	ID      uint    `json:"id" gorm:"primaryKey"`
	Kode    string  `json:"kode" gorm:"size:30;uniqueIndex;not null"`
	Nama    string  `json:"nama" gorm:"size:150;not null;index"`
	PIC     *string `json:"pic" gorm:"size:100"`
	Telepon string  `json:"telepon" gorm:"size:20"`
	// KerjasamaKurir menggantikan field Email lama — daftar kurir mitra
	// (dipisah koma) untuk pengiriman barang ke lokasi tujuan atas nama
	// supplier ini. TotalOrder/Rating di response API DIHITUNG dari sini
	// (lihat SupplierWithStats di repository), bukan kolom tersimpan.
	KerjasamaKurir string    `json:"kerjasama_kurir" gorm:"size:255"`
	Alamat         string    `json:"alamat" gorm:"size:255"`
	NPWP           *string   `json:"npwp" gorm:"size:25"`
	IsActive       bool      `json:"is_active" gorm:"not null;default:true"`
	IsProtected    bool      `json:"is_protected" gorm:"not null;default:false"` // dikunci super_admin — lihat internal/controller/supplier Protect()
	Catatan        string    `json:"catatan" gorm:"size:255"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (Supplier) TableName() string { return "supplier" }
