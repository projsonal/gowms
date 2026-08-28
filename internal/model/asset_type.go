package model

import "time"

type AssetType struct {
	ID    uint   `json:"id" gorm:"primaryKey"`
	Kode  string `json:"kode" gorm:"size:30;uniqueIndex;not null"`
	Label string `json:"label" gorm:"size:60;not null"`

	Color string `json:"color" gorm:"size:10;not null;default:'#6b7280'"`
	Abbr  string `json:"abbr" gorm:"size:6;not null"`

	HasKoordinat bool `json:"has_koordinat" gorm:"not null;default:true"`

	HasPort bool `json:"has_port" gorm:"not null;default:false"`

	IsSystem bool `json:"is_system" gorm:"not null;default:false"`
	Urutan   int  `json:"urutan" gorm:"not null;default:0"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (AssetType) TableName() string { return "asset_types" }
