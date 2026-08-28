package model

import "time"

type StockOpname struct {
	ID            uint              `json:"id" gorm:"primaryKey"`
	NomorOpname   string            `json:"nomor_opname" gorm:"size:30;uniqueIndex;not null"`
	GudangID      uint              `json:"gudang_id" gorm:"not null;index"`
	Gudang        *Gudang           `json:"gudang,omitempty" gorm:"foreignKey:GudangID"`
	Status        string            `json:"status" gorm:"size:20;not null;default:'draft';index"`
	Tanggal       time.Time         `json:"tanggal" gorm:"not null"`
	Catatan       string            `json:"catatan" gorm:"size:255"`
	DilakukanOleh *uint             `json:"dilakukan_oleh"`
	CompletedAt   *time.Time        `json:"completed_at"`
	Items         []StockOpnameItem `json:"items,omitempty" gorm:"foreignKey:StockOpnameID"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
}

func (StockOpname) TableName() string { return "stock_opname" }

type StockOpnameItem struct {
	ID            uint    `json:"id" gorm:"primaryKey"`
	StockOpnameID uint    `json:"stock_opname_id" gorm:"not null;index"`
	BarangID      uint    `json:"barang_id" gorm:"not null;index"`
	Barang        *Barang `json:"barang,omitempty" gorm:"foreignKey:BarangID"`
	StokSistem    int     `json:"stok_sistem" gorm:"not null;default:0"`
	StokFisik     int     `json:"stok_fisik" gorm:"not null;default:0"`
	Selisih       int     `json:"selisih" gorm:"not null;default:0"`
	Catatan       string  `json:"catatan" gorm:"size:255"`
}

func (StockOpnameItem) TableName() string { return "stock_opname_items" }

func (i *StockOpnameItem) HitungSelisih() {
	i.Selisih = i.StokFisik - i.StokSistem
}
