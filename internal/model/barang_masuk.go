package model

import "time"

type BarangMasuk struct {
	ID              uint              `json:"id" gorm:"primaryKey"`
	NomorPenerimaan string            `json:"nomor_penerimaan" gorm:"size:30;uniqueIndex;not null"` // BM-2026-0001
	GudangID        uint              `json:"gudang_id" gorm:"not null;index"`
	Gudang          *Gudang           `json:"gudang,omitempty" gorm:"foreignKey:GudangID"`
	Status          string            `json:"status" gorm:"size:20;not null;default:'draft';index"`
	Tanggal         time.Time         `json:"tanggal" gorm:"not null"`
	Catatan         string            `json:"catatan" gorm:"size:255"`
	DiterimaOleh    *uint             `json:"diterima_oleh"`
	CompletedAt     *time.Time        `json:"completed_at"`
	Items           []BarangMasukItem `json:"items,omitempty" gorm:"foreignKey:BarangMasukID"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
}

func (BarangMasuk) TableName() string { return "barang_masuk" }

type BarangMasukItem struct {
	ID            uint    `json:"id" gorm:"primaryKey"`
	BarangMasukID uint    `json:"barang_masuk_id" gorm:"not null;index"`
	BarangID      uint    `json:"barang_id" gorm:"not null;index"`
	Barang        *Barang `json:"barang,omitempty" gorm:"foreignKey:BarangID"`
	RakID         *uint   `json:"rak_id" gorm:"index"` // opsional: rak penempatan (dipilih manual operator)
	Rak           *Rak    `json:"rak,omitempty" gorm:"foreignKey:RakID"`
	Qty           int     `json:"qty" gorm:"not null"`
	HargaSatuan   int64   `json:"harga_satuan" gorm:"not null;default:0"`
}

func (BarangMasukItem) TableName() string { return "barang_masuk_items" }
