package barang_keluar

import (
	"github.com/projsonal/gowms/internal/model"
	"github.com/projsonal/gowms/pkg/utils"
)

type Filter struct {
	Status   string
	GudangID uint
}

type Repository interface {
	List(p utils.PaginationParams, f Filter) ([]model.BarangKeluar, int64, error)
	FindByID(id uint) (*model.BarangKeluar, error)
	FindByNomor(nomor string) (*model.BarangKeluar, error)
	Create(bk *model.BarangKeluar) error
	Update(bk *model.BarangKeluar, items []model.BarangKeluarItem) error
	Delete(id uint) error

	Complete(id uint, userID uint) (*model.BarangKeluar, error)
	Batalkan(id uint) (*model.BarangKeluar, error)

	CountByStatus(status string) (int64, error)
}
