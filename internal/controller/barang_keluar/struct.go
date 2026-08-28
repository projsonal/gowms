package barang_keluar

import (
	"time"

	barangRepo "github.com/projsonal/gowms/internal/repositories/barang"
	bkRepo "github.com/projsonal/gowms/internal/repositories/barang_keluar"
	gudangRepo "github.com/projsonal/gowms/internal/repositories/gudang"
	notificationRepo "github.com/projsonal/gowms/internal/repositories/notifikasi"
	"github.com/projsonal/gowms/internal/repositories/role"
	"github.com/projsonal/gowms/pkg/utils"
)

type Controller struct {
	repo       bkRepo.Repository
	barangRepo barangRepo.Repository
	gudangRepo gudangRepo.Repository
	roleRepo   role.Repository
	jwtSvc     *utils.JWTService
	notifRepo  notificationRepo.Repository
}

func New(repo bkRepo.Repository, barangRepo barangRepo.Repository, gudangRepo gudangRepo.Repository,
	roleRepo role.Repository, jwtSvc *utils.JWTService, notifRepo notificationRepo.Repository) *Controller {
	return &Controller{repo: repo, barangRepo: barangRepo, gudangRepo: gudangRepo, roleRepo: roleRepo, jwtSvc: jwtSvc, notifRepo: notifRepo}
}

type ItemRequest struct {
	BarangID uint `json:"barang_id" validate:"required"`
	Qty      int  `json:"qty" validate:"required,min=1"`
}

type BKRequest struct {
	GudangID uint `json:"gudang_id" validate:"required"`

	Tanggal   string        `json:"tanggal" validate:"required"`
	Keperluan string        `json:"keperluan" validate:"required,max=255"`
	Penerima  string        `json:"penerima" validate:"max=150"`
	Items     []ItemRequest `json:"items" validate:"required,min=1,dive"`
}

func parseTanggalHarian(raw string) (time.Time, error) {
	return time.Parse("2006-01-02", raw)
}

type CompleteBKRequest struct {
	Items []ItemSerialInput `json:"items"`
}

type ItemSerialInput struct {
	BarangKeluarItemID uint     `json:"barang_keluar_item_id" validate:"required"`
	SerialNumbers      []string `json:"serial_numbers"`
}

type SummaryResponse struct {
	TotalDokumen int64 `json:"total_dokumen"`
	Draft        int64 `json:"draft"`
	Selesai      int64 `json:"selesai"`
}
