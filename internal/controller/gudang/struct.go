package gudang

import (
	gudangRepo "github.com/projsonal/gowms/internal/repositories/gudang"
	"github.com/projsonal/gowms/internal/repositories/role"
	"github.com/projsonal/gowms/pkg/utils"
)

type Controller struct {
	repo     gudangRepo.Repository
	roleRepo role.Repository
	jwtSvc   *utils.JWTService
}

func New(repo gudangRepo.Repository, roleRepo role.Repository, jwtSvc *utils.JWTService) *Controller {
	return &Controller{repo: repo, roleRepo: roleRepo, jwtSvc: jwtSvc}
}

type KategoriRequest struct {
	Nama string `json:"nama" validate:"required,max=100"`
}

type SatuanRequest struct {
	Nama      string `json:"nama" validate:"required,max=50"`
	Singkatan string `json:"singkatan" validate:"required,max=10"`
}

type GudangRequest struct {
	Nama   string `json:"nama" validate:"required,max=100"`
	Alamat string `json:"alamat" validate:"max=255"`
}

type RakRequest struct {
	KodeRak   string `json:"kode_rak" validate:"required,max=20"`
	GudangID  uint   `json:"gudang_id" validate:"required"`
	Kapasitas int    `json:"kapasitas" validate:"required,min=1"`
}

type UpdateRakRequest struct {
	Kapasitas *int `json:"kapasitas" validate:"omitempty,min=1"`
}

type AdjustRakRequest struct {
	Delta int `json:"delta" validate:"required"`
}

type RakSummaryResponse struct {
	TotalGudang    int64 `json:"total_gudang"`
	TotalRak       int64 `json:"total_rak"`
	RakTerisiPenuh int64 `json:"rak_terisi_penuh"`
	RakKosong      int64 `json:"rak_kosong"`
}
