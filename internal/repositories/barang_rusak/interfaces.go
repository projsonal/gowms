package barang_rusak

import (
	"github.com/projsonal/gowms/internal/model"
	"github.com/projsonal/gowms/pkg/utils"
)

type Filter struct {
	Status string
}

type Repository interface {
	List(p utils.PaginationParams, f Filter) ([]model.BarangRusak, int64, error)
	FindByID(id uint) (*model.BarangRusak, error)
	Create(b *model.BarangRusak) error
	Update(b *model.BarangRusak) error
	Delete(id uint) error

	CountByStatus(status string) (int64, error)
}
