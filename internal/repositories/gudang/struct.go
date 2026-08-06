package gudang

import "gorm.io/gorm"

type repository struct {
	db *gorm.DB
}

// membuat baru instance Repository gudang berbasis GORM.
func New(db *gorm.DB) Repository {
	return &repository{db: db}
}
