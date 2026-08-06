package barang

import (
	"github.com/projsonal/gowms/internal/model"
	"github.com/projsonal/gowms/pkg/utils"
)

type Filter struct {
	KategoriID  uint
	SatuanID    uint
	StokMenipis bool // true = hanya tampilkan barang dengan stok <= stok_minimum (dan stok_minimum > 0)
	OnlyActive  bool // true = sembunyikan barang yang IsActive=false (didiskontinu)
}

type Repository interface {
	List(p utils.PaginationParams, f Filter) ([]model.Barang, int64, error)
	FindByID(id uint) (*model.Barang, error)
	FindByKode(kode string) (*model.Barang, error)
	Create(b *model.Barang) error
	Update(b *model.Barang) error
	Delete(id uint) error

	AdjustStok(id uint, delta int) (*model.Barang, error)

	CountAll() (int64, error)
	CountStokMenipis() (int64, error)
	SumNilaiInventaris() (int64, error)
}
