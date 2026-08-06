package users

import "gorm.io/gorm"

type repository struct {
	db *gorm.DB
}

// New membuat instance Repository user berbasis GORM.
func New(db *gorm.DB) Repository {
	return &repository{db: db}
}
