package model

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

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
	ID   uint   `json:"id" gorm:"primaryKey"`
	Nama string `json:"nama" gorm:"size:100;not null"`

	Kode   string `json:"kode" gorm:"size:20;uniqueIndex"`
	Alamat string `json:"alamat" gorm:"size:255"`

	Tipe string `json:"tipe" gorm:"size:10;not null;default:'cabang';index"`

	PIC string `json:"pic" gorm:"size:100"`

	Telepon   string `json:"telepon" gorm:"size:20"`
	Kapasitas int    `json:"kapasitas" gorm:"not null;default:0"`

	Latitude    *float64  `json:"latitude"`
	Longitude   *float64  `json:"longitude"`
	IsProtected bool      `json:"is_protected" gorm:"not null;default:false"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	UnitTersedia int64 `json:"-" gorm:"-"`
	SkuTersedia  int64 `json:"-" gorm:"-"`
}

func (Gudang) TableName() string { return "gudangs" }

func (g Gudang) MarshalJSON() ([]byte, error) {
	type Alias Gudang
	return json.Marshal(struct {
		Alias
		UnitTersedia int64 `json:"unit_tersedia"`
		SkuTersedia  int64 `json:"sku_tersedia"`
	}{
		Alias:        Alias(g),
		UnitTersedia: g.UnitTersedia,
		SkuTersedia:  g.SkuTersedia,
	})
}
