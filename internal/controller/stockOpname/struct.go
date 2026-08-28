package stock_opname

import (
	"time"

	barangRepo "github.com/projsonal/gowms/internal/repositories/barang"
	gudangRepo "github.com/projsonal/gowms/internal/repositories/gudang"
	notificationRepo "github.com/projsonal/gowms/internal/repositories/notifikasi"
	"github.com/projsonal/gowms/internal/repositories/role"
	soRepo "github.com/projsonal/gowms/internal/repositories/stockOpname"
	"github.com/projsonal/gowms/pkg/utils"
)

type Controller struct {
	repo       soRepo.Repository
	barangRepo barangRepo.Repository
	gudangRepo gudangRepo.Repository
	roleRepo   role.Repository
	jwtSvc     *utils.JWTService
	notifRepo  notificationRepo.Repository
}

func New(repo soRepo.Repository, barangRepo barangRepo.Repository, gudangRepo gudangRepo.Repository,
	roleRepo role.Repository, jwtSvc *utils.JWTService, notifRepo notificationRepo.Repository) *Controller {
	return &Controller{repo: repo, barangRepo: barangRepo, gudangRepo: gudangRepo, roleRepo: roleRepo, jwtSvc: jwtSvc, notifRepo: notifRepo}
}

type ItemRequest struct {
	BarangID  uint   `json:"barang_id" validate:"required"`
	StokFisik int    `json:"stok_fisik" validate:"min=0"`
	Catatan   string `json:"catatan" validate:"max=255"`
}

type SORequest struct {
	GudangID uint `json:"gudang_id" validate:"required"`

	Tanggal string        `json:"tanggal" validate:"required"`
	Catatan string        `json:"catatan" validate:"max=255"`
	Items   []ItemRequest `json:"items" validate:"required,min=1,dive"`
}

func parseTanggalHarian(raw string) (time.Time, error) {
	return time.Parse("2006-01-02", raw)
}

type SummaryResponse struct {
	TotalDokumen int64 `json:"total_dokumen"`
	Draft        int64 `json:"draft"`
	Selesai      int64 `json:"selesai"`
}
