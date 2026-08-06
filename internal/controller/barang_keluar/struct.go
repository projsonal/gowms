package barang_keluar

import (
	"time"

	barangRepo "github.com/projsonal/gostock/internal/repositories/barang"
	bkRepo "github.com/projsonal/gostock/internal/repositories/barang_keluar"
	gudangRepo "github.com/projsonal/gostock/internal/repositories/gudang"
	"github.com/projsonal/gostock/internal/repositories/role"
	"github.com/projsonal/gostock/pkg/utils"
)

type Controller struct {
	repo       bkRepo.Repository
	barangRepo barangRepo.Repository
	gudangRepo gudangRepo.Repository
	roleRepo   role.Repository
	jwtSvc     *utils.JWTService
}

func New(repo bkRepo.Repository, barangRepo barangRepo.Repository, gudangRepo gudangRepo.Repository,
	roleRepo role.Repository, jwtSvc *utils.JWTService) *Controller {
	return &Controller{repo: repo, barangRepo: barangRepo, gudangRepo: gudangRepo, roleRepo: roleRepo, jwtSvc: jwtSvc}
}

type ItemRequest struct {
	BarangID uint  `json:"barang_id" validate:"required"`
	RakID    *uint `json:"rak_id"`
	Qty      int   `json:"qty" validate:"required,min=1"`
}

type BKRequest struct {
	GudangID  uint          `json:"gudang_id" validate:"required"`
	Tanggal   time.Time     `json:"tanggal" validate:"required"`
	Keperluan string        `json:"keperluan" validate:"required,max=255"`
	Penerima  string        `json:"penerima" validate:"max=150"`
	Items     []ItemRequest `json:"items" validate:"required,min=1,dive"`
}

type SummaryResponse struct {
	TotalDokumen int64 `json:"total_dokumen"`
	Draft        int64 `json:"draft"`
	Selesai      int64 `json:"selesai"`
}
