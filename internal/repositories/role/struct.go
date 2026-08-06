package role

import "gorm.io/gorm"

type repository struct {
	db *gorm.DB
}

// New membuat instance Repository role/permission berbasis GORM.
func New(db *gorm.DB) Repository {
	return &repository{db: db}
}
