package model

import "time"

type Supplier struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Kode      string    `json:"kode" gorm:"size:30;uniqueIndex;not null"`
	Nama      string    `json:"nama" gorm:"size:150;not null;index"`
	PIC       *string   `json:"pic" gorm:"size:100"`
	Telepon   string    `json:"telepon" gorm:"size:20"`
	Email     string    `json:"email" gorm:"size:100"`
	Alamat    string    `json:"alamat" gorm:"size:255"`
	NPWP      *string   `json:"npwp" gorm:"size:25"`
	IsActive  bool      `json:"is_active" gorm:"not null;default:true"`
	Catatan   string    `json:"catatan" gorm:"size:255"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Supplier) TableName() string { return "supplier" }
