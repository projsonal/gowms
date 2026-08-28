package model

import (
	"time"

	"gorm.io/gorm"
)

type AssetPort struct {
	ID         uint `json:"id" gorm:"primaryKey"`
	AssetID    uint `json:"asset_id" gorm:"not null;index:idx_asset_port,unique"`
	PortNumber int  `json:"port_number" gorm:"not null;index:idx_asset_port,unique"`

	Status string `json:"status" gorm:"size:10;not null;default:'kosong'"`

	ChildAssetID *uint  `json:"child_asset_id" gorm:"index"`
	ChildAsset   *Asset `json:"child_asset,omitempty" gorm:"foreignKey:ChildAssetID"`

	CustomerName  string `json:"customer_name" gorm:"size:150"`
	CustomerPhone string `json:"customer_phone" gorm:"size:20"`
	Keterangan    string `json:"keterangan" gorm:"size:255"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (AssetPort) TableName() string { return "asset_ports" }
