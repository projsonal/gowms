package model

import "time"

type Barang struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	KodeBarang  string    `json:"kode_barang" gorm:"size:30;uniqueIndex;not null"` // SKU, mis: "BRG-0001"
	Nama        string    `json:"nama" gorm:"size:150;not null;index"`
	KategoriID  uint      `json:"kategori_id" gorm:"not null;index"`
	Kategori    *Kategori `json:"kategori,omitempty" gorm:"foreignKey:KategoriID"`
	SatuanID    uint      `json:"satuan_id" gorm:"not null;index"`
	Satuan      *Satuan   `json:"satuan,omitempty" gorm:"foreignKey:SatuanID"`
	HargaBeli   int64     `json:"harga_beli" gorm:"not null;default:0"`   // rupiah/unit
	StokMinimum int       `json:"stok_minimum" gorm:"not null;default:0"` // reorder point; 0 = tidak dipantau
	Stok        int       `json:"stok" gorm:"not null;default:0"`         // agregat total, lihat AdjustStok()
	IsActive    bool      `json:"is_active" gorm:"not null;default:true"` // nonaktif = didiskontinu, tetap tampil di histori
	Deskripsi   string    `json:"deskripsi" gorm:"size:255"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (Barang) TableName() string { return "barang" }

func (b *Barang) IsStokMenipis() bool {
	return b.StokMinimum > 0 && b.Stok <= b.StokMinimum
}

func (b *Barang) NilaiInventaris() int64 {
	return int64(b.Stok) * b.HargaBeli
}
