package model

type assets struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	BarangID    uint   `json:"barang_id" gorm:"index;not null"`
	labelbrng   string `json:"label_barang"`
	nama_barang string `json:"nama_barang"`
}
