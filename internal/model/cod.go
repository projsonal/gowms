package model

import "time"

type CodTransaction struct {
	ID           uint        `json:"id" gorm:"primaryKey"`
	Kode         string      `json:"kode" gorm:"size:30;uniqueIndex;not null"`
	PengirimanID *uint       `json:"pengiriman_id" gorm:"index"`
	Pengiriman   *Pengiriman `json:"pengiriman,omitempty" gorm:"foreignKey:PengirimanID"`
	Pelanggan    string      `json:"pelanggan" gorm:"size:150;not null"`
	Nominal      int64       `json:"nominal" gorm:"not null;default:0"`
	Kurir        string      `json:"kurir" gorm:"size:100"`
	Tanggal      time.Time   `json:"tanggal" gorm:"not null"`
	Status       string      `json:"status" gorm:"size:20;not null;default:'menunggu';index"`
	IsProtected  bool        `json:"is_protected" gorm:"not null;default:false"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
}

func (CodTransaction) TableName() string { return "cod_transactions" }
