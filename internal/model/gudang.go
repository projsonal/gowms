package model

import "time"

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
	ID        uint      `json:"id" gorm:"primaryKey"`
	Nama      string    `json:"nama" gorm:"size:100;not null"`
	Alamat    string    `json:"alamat" gorm:"size:255"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Raks      []Rak     `json:"raks,omitempty" gorm:"foreignKey:GudangID"`
}

func (Gudang) TableName() string { return "gudangs" }

type Rak struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	KodeRak   string    `json:"kode_rak" gorm:"size:20;uniqueIndex;not null"` // A2-R05
	GudangID  uint      `json:"gudang_id" gorm:"not null;index"`
	Gudang    *Gudang   `json:"gudang,omitempty" gorm:"foreignKey:GudangID"`
	Kapasitas int       `json:"kapasitas" gorm:"not null;default:0"`
	Terisi    int       `json:"terisi" gorm:"not null;default:0"`
	Status    string    `json:"status" gorm:"size:20;default:'kosong'"` // kosong|terisi_sebagian|penuh
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Rak) TableName() string { return "raks" }

func (r *Rak) RecalculateStatus() {
	switch {
	case r.Terisi <= 0:
		r.Status = "kosong"
	case r.Terisi >= r.Kapasitas:
		r.Status = "penuh"
	default:
		r.Status = "terisi_sebagian"
	}
}
