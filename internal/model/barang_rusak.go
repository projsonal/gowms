package model

import (
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type BarangRusak struct {
	ID uint `json:"id" gorm:"primaryKey"`

	BarangID *uint   `json:"barang_id"`
	Barang   *Barang `json:"barang,omitempty" gorm:"foreignKey:BarangID"`

	LabelBarang string `json:"label_barang" gorm:"size:60;not null;index"`
	NamaBarang  string `json:"nama_barang" gorm:"size:150;not null"`

	SerialNumber string `json:"serial_number" gorm:"size:100"`
	Keterangan   string `json:"keterangan" gorm:"size:500"`

	FotoData        []byte `json:"-" gorm:"type:bytea"`
	FotoContentType string `json:"-" gorm:"size:100"`

	JenisBarang string `json:"jenis_barang" gorm:"size:10"`

	Status string `json:"status" gorm:"size:20;not null;default:'pengecekan';index"`

	DilaporkanOleh uint       `json:"dilaporkan_oleh" gorm:"not null"`
	Pelapor        *User      `json:"pelapor,omitempty" gorm:"foreignKey:DilaporkanOleh"`
	DicekOleh      *uint      `json:"dicek_oleh"`
	Pemeriksa      *User      `json:"pemeriksa,omitempty" gorm:"foreignKey:DicekOleh"`
	DicekPada      *time.Time `json:"dicek_pada"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (BarangRusak) TableName() string { return "barang_rusak" }

func (b BarangRusak) FotoURL() string {
	if len(b.FotoData) == 0 {
		return ""
	}
	return fmt.Sprintf("/barang-rusak/%d/foto?v=%d", b.ID, b.UpdatedAt.Unix())
}

func (b BarangRusak) MarshalJSON() ([]byte, error) {
	type Alias BarangRusak
	return json.Marshal(struct {
		Alias
		FotoURL string `json:"foto_url"`
	}{
		Alias:   Alias(b),
		FotoURL: b.FotoURL(),
	})
}
