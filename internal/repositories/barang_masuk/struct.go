package barang_masuk

import (
	"gorm.io/gorm"

	poRepo "github.com/projsonal/gostock/internal/repositories/po"
)

type repository struct {
	db     *gorm.DB
	poRepo poRepo.Repository // dipakai TambahPenerimaan saat Complete, kalau BM ini berasal dari PO
}

func New(db *gorm.DB, poRepo poRepo.Repository) Repository {
	return &repository{db: db, poRepo: poRepo}
}
