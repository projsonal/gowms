package model

import "time"

type AssetHistory struct {
	ID      uint `json:"id" gorm:"primaryKey"`
	AssetID uint `json:"asset_id" gorm:"not null;index"`

	EventType string `json:"event_type" gorm:"size:20;not null;index"`
	FieldLama string `json:"field_lama" gorm:"size:255"`
	FieldBaru string `json:"field_baru" gorm:"size:255"`
	Catatan   string `json:"catatan" gorm:"size:255"`

	UserID   *uint  `json:"user_id"`
	UserNama string `json:"user_nama" gorm:"size:150"`

	CreatedAt time.Time `json:"created_at"`
}

func (AssetHistory) TableName() string { return "asset_histories" }
