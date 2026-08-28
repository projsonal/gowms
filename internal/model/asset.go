package model

import (
	"time"

	"gorm.io/gorm"
)

type Asset struct {
	ID   uint   `json:"id" gorm:"primaryKey"`
	Nama string `json:"nama" gorm:"size:150;not null"`

	JenisAset string `json:"jenis_aset" gorm:"size:20;not null;index"`

	GudangID uint    `json:"gudang_id" gorm:"not null;index"`
	Gudang   *Gudang `json:"gudang,omitempty" gorm:"foreignKey:GudangID"`

	LabelRSD string `json:"label_rsd" gorm:"size:40;index"`

	KodeBA string `json:"kode_ba" gorm:"size:20;index"`

	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`

	ParentAssetID *uint  `json:"parent_asset_id" gorm:"index"`
	Parent        *Asset `json:"parent,omitempty" gorm:"foreignKey:ParentAssetID"`

	JumlahPort int `json:"jumlah_port" gorm:"not null;default:0"`

	Status     string `json:"status" gorm:"size:20;not null;default:'aktif';index"`
	Keterangan string `json:"keterangan" gorm:"size:500"`

	Merek string `json:"merek" gorm:"size:100"`
	Tipe  string `json:"tipe" gorm:"size:100"`

	BarangID *uint   `json:"barang_id" gorm:"index"`
	Barang   *Barang `json:"barang,omitempty" gorm:"foreignKey:BarangID"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (Asset) TableName() string { return "assets" }

func JenisAsetPunyaKoordinat(jenisAset string) bool {
	return jenisAset != "transportasi"
}

var jenisIndukValid = map[string]map[string]bool{
	"odc":   {"olt": true},
	"odp":   {"olt": true, "odc": true},
	"ont":   {"olt": true, "odc": true, "odp": true},
	"tiang": {"olt": true, "odc": true, "odp": true, "tiang": true},
	"modem": {"olt": true, "odc": true, "odp": true, "ont": true},
}

func JenisIndukValid(childJenis, parentJenis string) bool {
	allowed, ok := jenisIndukValid[childJenis]
	if !ok {
		return true
	}
	return allowed[parentJenis]
}
