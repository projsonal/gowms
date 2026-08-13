package model

type asset struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	BarangID    uint   `json:"barang_id" gorm:"index;not null"`
	LabelBarang string `json:"label_barang"`
	NamaBarang  string `json:"nama_barang"`
}
