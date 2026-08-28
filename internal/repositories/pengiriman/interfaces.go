package pengiriman

import (
	"time"

	"github.com/projsonal/gowms/internal/model"
	"github.com/projsonal/gowms/pkg/utils"
)

type Filter struct {
	Status       string
	GudangAsalID uint
	Jenis        string
}

type Repository interface {
	List(p utils.PaginationParams, f Filter) ([]model.Pengiriman, int64, error)
	FindByID(id uint) (*model.Pengiriman, error)
	FindByNomor(nomor string) (*model.Pengiriman, error)
	Create(pg *model.Pengiriman) error
	Update(pg *model.Pengiriman) error
	Delete(id uint) error

	Jadwalkan(id uint, namaKurir, teleponKurir string, estimasiTiba *time.Time) (*model.Pengiriman, error)
	Mulai(id uint) (*model.Pengiriman, error)

	RecordLocation(id uint, lat, lng float64, kecepatan *float64, recordedAt time.Time) (*model.Pengiriman, error)
	ListTrackingPoints(pengirimanID uint, limit int) ([]model.PengirimanTrackingPoint, error)

	Selesaikan(id uint, catatan string) (*model.Pengiriman, error)
	Batalkan(id uint) (*model.Pengiriman, error)
	SetProtected(id uint, protected bool) (*model.Pengiriman, error)

	CountByStatus(status string) (int64, error)
}
