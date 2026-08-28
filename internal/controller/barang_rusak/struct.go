package barang_rusak

import (
	notificationRepo "github.com/projsonal/gowms/internal/repositories/notifikasi"

	barangRepo "github.com/projsonal/gowms/internal/repositories/barang"
	barangRusakRepo "github.com/projsonal/gowms/internal/repositories/barang_rusak"
	"github.com/projsonal/gowms/internal/repositories/role"
	"github.com/projsonal/gowms/pkg/constant"
	"github.com/projsonal/gowms/pkg/utils"
)

const Module = constant.ModuleBarangRusak

type Controller struct {
	repo       barangRusakRepo.Repository
	barangRepo barangRepo.Repository
	roleRepo   role.Repository
	jwtSvc     *utils.JWTService
	notifRepo  notificationRepo.Repository
}

func New(repo barangRusakRepo.Repository, barangRepo barangRepo.Repository, roleRepo role.Repository, jwtSvc *utils.JWTService, notifRepo notificationRepo.Repository) *Controller {
	return &Controller{repo: repo, barangRepo: barangRepo, roleRepo: roleRepo, jwtSvc: jwtSvc, notifRepo: notifRepo}
}

type BarangRusakRequest struct {
	BarangID     *uint  `json:"barang_id" validate:"omitempty"`
	LabelBarang  string `json:"label_barang" validate:"required,max=60"`
	NamaBarang   string `json:"nama_barang" validate:"required,max=150"`
	SerialNumber string `json:"serial_number" validate:"max=100"`
	Keterangan   string `json:"keterangan" validate:"max=500"`
}

type InspeksiRequest struct {
	JenisBarang string `json:"jenis_barang" validate:"required,oneof=retur rusak"`
}

type SummaryResponse struct {
	Pengecekan int64 `json:"pengecekan"`
	Retur      int64 `json:"retur"`
	Rusak      int64 `json:"rusak"`
	Total      int64 `json:"total"`
}
