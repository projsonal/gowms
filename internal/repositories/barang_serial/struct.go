package barang_serial

import "gorm.io/gorm"

type repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) Repository {
	return &repository{db: db}
}
