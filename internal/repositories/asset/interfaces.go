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

type Repository interface {
	List(p utils.PaginationParams, f Filter) ([]model.Asset, int64, error)
	FindByID(id uint) (*model.Asset, error)
	Create(a *model.Asset) error
	Update(a *model.Asset) error
	Delete(id uint) error

	// NextRSDNumber — nomor urut berikutnya untuk label RSD di gudang
	// tertentu (reset per gudang, dihitung dari total aset berkoordinat
	// yang sudah ada di gudang itu + 1).
	NextRSDNumber(gudangID uint) (int, error)
	// NextBANumber — nomor urut berikutnya untuk kode BA (global, lintas
	// gudang) khusus aset transportasi.
	NextBANumber() (int, error)

	CountByJenis(jenisAset string) (int64, error)
}
