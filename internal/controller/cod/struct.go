package cod

import (
	codRepo "github.com/projsonal/gowms/internal/repositories/cod"
	"github.com/projsonal/gowms/internal/repositories/role"
	"github.com/projsonal/gowms/pkg/utils"
)

type Controller struct {
	repo     codRepo.Repository
	roleRepo role.Repository
	jwtSvc   *utils.JWTService
}

func New(repo codRepo.Repository, roleRepo role.Repository, jwtSvc *utils.JWTService) *Controller {
	return &Controller{repo: repo, roleRepo: roleRepo, jwtSvc: jwtSvc}
}

type CodRequest struct {
	Kode      string `json:"kode" validate:"required,max=30"`
	Pelanggan string `json:"pelanggan" validate:"required,max=150"`
	Nominal   int64  `json:"nominal" validate:"required,min=1"`
	Kurir     string `json:"kurir" validate:"max=100"`
	Tanggal   string `json:"tanggal" validate:"required"`
	Status    string `json:"status" validate:"required,oneof=menunggu lunas bermasalah"`
}

type ProtectRequest struct {
	IsProtected *bool `json:"is_protected" validate:"required"`
}

type SummaryResponse struct {
	Total        int64 `json:"total"`
	Lunas        int64 `json:"lunas"`
	Menunggu     int64 `json:"menunggu"`
	Bermasalah   int64 `json:"bermasalah"`
	TotalNominal int64 `json:"total_nominal"`
}
