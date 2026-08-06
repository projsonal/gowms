package supplier

import (
	"github.com/projsonal/gostock/internal/model"
	"github.com/projsonal/gostock/pkg/utils"
)

type Filter struct {
	OnlyActive bool
}

type Repository interface {
	List(p utils.PaginationParams, f Filter) ([]model.Supplier, int64, error)
	FindByID(id uint) (*model.Supplier, error)
	FindByKode(kode string) (*model.Supplier, error)
	Create(s *model.Supplier) error
	Update(s *model.Supplier) error
	Delete(id uint) error

	CountAll() (int64, error)
	CountActive() (int64, error)
}
