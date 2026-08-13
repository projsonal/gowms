package assetgudang

import (
	assetRepo "github.com/projsonal/gowms/internal/repositories/asset"
	gudangRepo "github.com/projsonal/gowms/internal/repositories/gudang"
	"github.com/projsonal/gowms/internal/repositories/role"
	"github.com/projsonal/gowms/pkg/constant"
	"github.com/projsonal/gowms/pkg/utils"
)

const Module = constant.ModuleAsetGudang

type Controller struct {
	repo       assetRepo.Repository
	gudangRepo gudangRepo.Repository
	roleRepo   role.Repository
	jwtSvc     *utils.JWTService
}

func New(repo assetRepo.Repository, gudangRepo gudangRepo.Repository, roleRepo role.Repository, jwtSvc *utils.JWTService) *Controller {
	return &Controller{repo: repo, gudangRepo: gudangRepo, roleRepo: roleRepo, jwtSvc: jwtSvc}
}

// AssetRequest — payload Tambah/Ubah Aset. Latitude & Longitude WAJIB
// diisi untuk jenis_aset selain "transportasi" (lihat Controller.Create).
type AssetRequest struct {
	Nama       string   `json:"nama" validate:"required,max=150"`
	JenisAset  string   `json:"jenis_aset" validate:"required,oneof=tiang odc ont odp olt transportasi"`
	GudangID   uint     `json:"gudang_id" validate:"required"`
	Latitude   *float64 `json:"latitude" validate:"omitempty,min=-90,max=90"`
	Longitude  *float64 `json:"longitude" validate:"omitempty,min=-180,max=180"`
	Keterangan string   `json:"keterangan" validate:"max=500"`
}

// UpdateStatusRequest — payload PATCH /aset/:id/status untuk menandai
// kondisi aset (mis. setelah pemeriksaan lapangan).
type UpdateStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=aktif rusak nonaktif"`
}

type SummaryResponse struct {
	Tiang        int64 `json:"tiang"`
	Odc          int64 `json:"odc"`
	Ont          int64 `json:"ont"`
	Odp          int64 `json:"odp"`
	Olt          int64 `json:"olt"`
	Transportasi int64 `json:"transportasi"`
	Total        int64 `json:"total"`
}
