package barang_serial

import (
	"github.com/projsonal/gowms/internal/model"
	"github.com/projsonal/gowms/pkg/utils"
)

type Filter struct {
	BarangID uint
	GudangID uint
	Status   string

	BarangMasukItemID  uint
	BarangKeluarItemID uint
}

type Repository interface {
	List(p utils.PaginationParams, f Filter) ([]model.BarangSerial, int64, error)
	FindByID(id uint) (*model.BarangSerial, error)

	FindBySerial(serial string) (*model.BarangSerial, error)
	CountByBarang(barangID uint) (tersedia int64, terpasang int64, rusak int64, err error)

	Create(barangID, gudangID uint, serialNumber, catatan string) (*model.BarangSerial, error)

	RiwayatDokumen(s *model.BarangSerial) (nomorMasuk string, nomorKeluar string, err error)

	UpdateStatusManual(id uint, status string, catatan string) (*model.BarangSerial, error)
	Delete(id uint) error
}
