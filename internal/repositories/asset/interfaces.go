package asset

import (
	"github.com/projsonal/gowms/internal/model"
	"github.com/projsonal/gowms/pkg/utils"
)

type Filter struct {
	JenisAset string
	GudangID  uint
	Status    string
}

type MapRow struct {
	ID              uint
	Nama            string
	JenisAset       string
	LabelRSD        string
	Latitude        float64
	Longitude       float64
	Status          string
	GudangID        uint
	GudangNama      string
	GudangKode      string
	GudangTipe      string
	GudangLatitude  *float64
	GudangLongitude *float64
	ParentAssetID   *uint
	ParentLatitude  *float64
	ParentLongitude *float64
	JumlahPort      int
	PortTerisi      int64

	Merek string
	Tipe  string

	KodeBarang string
}

type Repository interface {
	List(p utils.PaginationParams, f Filter) ([]model.Asset, int64, error)
	FindByID(id uint) (*model.Asset, error)

	ListForMap(f Filter, tipeGudang string) ([]MapRow, error)
	Create(a *model.Asset) error
	Update(a *model.Asset) error
	Delete(id uint) error

	NextRSDNumber(gudangID uint) (int, error)

	NextBANumber() (int, error)

	CountByJenis(jenisAset string) (int64, error)
}
