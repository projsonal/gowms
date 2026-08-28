package model

import (
	"time"

	"gorm.io/gorm"
)

type BarangSerial struct {
	ID       uint    `json:"id" gorm:"primaryKey"`
	BarangID uint    `json:"barang_id" gorm:"not null;index"`
	Barang   *Barang `json:"barang,omitempty" gorm:"foreignKey:BarangID"`

	SerialNumber string `json:"serial_number" gorm:"size:100;not null;uniqueIndex"`

	Status string `json:"status" gorm:"size:20;not null;default:'tersedia';index"`

	GudangID *uint   `json:"gudang_id" gorm:"index"`
	Gudang   *Gudang `json:"gudang,omitempty" gorm:"foreignKey:GudangID"`

	BarangMasukItemID  *uint             `json:"barang_masuk_item_id" gorm:"index"`
	BarangMasukItem    *BarangMasukItem  `json:"barang_masuk_item,omitempty" gorm:"foreignKey:BarangMasukItemID"`
	BarangKeluarItemID *uint             `json:"barang_keluar_item_id" gorm:"index"`
	BarangKeluarItem   *BarangKeluarItem `json:"barang_keluar_item,omitempty" gorm:"foreignKey:BarangKeluarItemID"`

	Catatan string `json:"catatan" gorm:"size:255"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (BarangSerial) TableName() string { return "barang_serials" }
