package model

import "time"

type BarangStokGudang struct {
	ID       uint    `json:"id" gorm:"primaryKey"`
	BarangID uint    `json:"barang_id" gorm:"not null;uniqueIndex:idx_barang_gudang"`
	Barang   *Barang `json:"barang,omitempty" gorm:"foreignKey:BarangID"`
	GudangID uint    `json:"gudang_id" gorm:"not null;uniqueIndex:idx_barang_gudang"`
	Gudang   *Gudang `json:"gudang,omitempty" gorm:"foreignKey:GudangID"`
	Stok     int     `json:"stok" gorm:"not null;default:0"`

	UpdatedAt time.Time `json:"updated_at"`
}

func (BarangStokGudang) TableName() string { return "barang_stok_gudang" }
