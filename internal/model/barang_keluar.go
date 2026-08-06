package model

import "time"

type BarangKeluar struct {
	ID               uint               `json:"id" gorm:"primaryKey"`
	NomorPengeluaran string             `json:"nomor_pengeluaran" gorm:"size:30;uniqueIndex;not null"` // BK-2026-0001
	GudangID         uint               `json:"gudang_id" gorm:"not null;index"`
	Gudang           *Gudang            `json:"gudang,omitempty" gorm:"foreignKey:GudangID"`
	Status           string             `json:"status" gorm:"size:20;not null;default:'draft';index"`
	Tanggal          time.Time          `json:"tanggal" gorm:"not null"`
	Keperluan        string             `json:"keperluan" gorm:"size:255"` // tujuan/alasan pengeluaran
	Penerima         string             `json:"penerima" gorm:"size:150"`  // pihak yang menerima barang
	DikeluarkanOleh  *uint              `json:"dikeluarkan_oleh"`
	CompletedAt      *time.Time         `json:"completed_at"`
	Items            []BarangKeluarItem `json:"items,omitempty" gorm:"foreignKey:BarangKeluarID"`
	CreatedAt        time.Time          `json:"created_at"`
	UpdatedAt        time.Time          `json:"updated_at"`
}

func (BarangKeluar) TableName() string { return "barang_keluar" }

type BarangKeluarItem struct {
	ID             uint    `json:"id" gorm:"primaryKey"`
	BarangKeluarID uint    `json:"barang_keluar_id" gorm:"not null;index"`
	BarangID       uint    `json:"barang_id" gorm:"not null;index"`
	Barang         *Barang `json:"barang,omitempty" gorm:"foreignKey:BarangID"`
	RakID          *uint   `json:"rak_id" gorm:"index"` // opsional: rak asal pengambilan
	Rak            *Rak    `json:"rak,omitempty" gorm:"foreignKey:RakID"`
	Qty            int     `json:"qty" gorm:"not null"`
}

func (BarangKeluarItem) TableName() string { return "barang_keluar_items" }
