package model

import (
	"time"

	"gorm.io/gorm"
)

type Barang struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	KodeBarang  string    `json:"kode_barang" gorm:"size:30;uniqueIndex;not null"`
	Nama        string    `json:"nama" gorm:"size:150;not null;index"`
	KategoriID  uint      `json:"kategori_id" gorm:"not null;index"`
	Kategori    *Kategori `json:"kategori,omitempty" gorm:"foreignKey:KategoriID"`
	SatuanID    uint      `json:"satuan_id" gorm:"not null;index"`
	Satuan      *Satuan   `json:"satuan,omitempty" gorm:"foreignKey:SatuanID"`
	HargaBeli   int64     `json:"harga_beli" gorm:"not null;default:0"`
	StokMinimum int       `json:"stok_minimum" gorm:"not null;default:0"`
	Stok        int       `json:"stok" gorm:"not null;default:0"`

	BeratGram   *int   `json:"berat_gram"`
	IsActive    bool   `json:"is_active" gorm:"not null;default:true"`
	IsProtected bool   `json:"is_protected" gorm:"not null;default:false"`
	Deskripsi   string `json:"deskripsi" gorm:"size:255"`

	Merek string `json:"merek" gorm:"size:100"`
	Tipe  string `json:"tipe" gorm:"size:100"`

	IsSerialized bool `json:"is_serialized" gorm:"not null;default:false"`

	ApprovalStatus string `json:"approval_status" gorm:"size:20;not null;default:'disetujui';index"`
	DiajukanOleh   *uint  `json:"diajukan_oleh"`
	DisetujuiOleh  *uint  `json:"disetujui_oleh"`

	DidelegasikanKe *uint      `json:"didelegasikan_ke"`
	Didelegasikan   *User      `json:"didelegasikan,omitempty" gorm:"foreignKey:DidelegasikanKe"`
	CatatanApproval string     `json:"catatan_approval" gorm:"size:255"`
	DireviewPada    *time.Time `json:"direview_pada"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (Barang) TableName() string { return "barang" }

func (b *Barang) IsStokMenipis() bool {
	return b.StokMinimum > 0 && b.Stok <= b.StokMinimum
}

func (b *Barang) NilaiInventaris() int64 {
	return int64(b.Stok) * b.HargaBeli
}
