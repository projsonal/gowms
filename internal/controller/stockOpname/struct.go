// Package stock_opname mengimplementasikan layar "Stock Opname": pencocokan
// stok sistem vs hasil hitung fisik manual di lapangan (tanpa alat IoT).
package stock_opname

import (
	"time"

	barangRepo "github.com/projsonal/gostock/internal/repositories/barang"
	gudangRepo "github.com/projsonal/gostock/internal/repositories/gudang"
	"github.com/projsonal/gostock/internal/repositories/role"
	soRepo "github.com/projsonal/gostock/internal/repositories/stockOpname"
	"github.com/projsonal/gostock/pkg/utils"
)

type Controller struct {
	repo       soRepo.Repository
	barangRepo barangRepo.Repository
	gudangRepo gudangRepo.Repository
	roleRepo   role.Repository
	jwtSvc     *utils.JWTService
}

func New(repo soRepo.Repository, barangRepo barangRepo.Repository, gudangRepo gudangRepo.Repository,
	roleRepo role.Repository, jwtSvc *utils.JWTService) *Controller {
	return &Controller{repo: repo, barangRepo: barangRepo, gudangRepo: gudangRepo, roleRepo: roleRepo, jwtSvc: jwtSvc}
}

// ---- DTO ----

type ItemRequest struct {
	BarangID  uint   `json:"barang_id" validate:"required"`
	RakID     *uint  `json:"rak_id"`
	StokFisik int    `json:"stok_fisik" validate:"min=0"`
	Catatan   string `json:"catatan" validate:"max=255"`
}

type SORequest struct {
	GudangID uint          `json:"gudang_id" validate:"required"`
	Tanggal  time.Time     `json:"tanggal" validate:"required"`
	Catatan  string        `json:"catatan" validate:"max=255"`
	Items    []ItemRequest `json:"items" validate:"required,min=1,dive"`
}

type SummaryResponse struct {
	TotalDokumen int64 `json:"total_dokumen"`
	Draft        int64 `json:"draft"`
	Selesai      int64 `json:"selesai"`
}
