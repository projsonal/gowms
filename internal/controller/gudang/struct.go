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
	Nama string `json:"nama" validate:"required,max=100"`

	Kode string `json:"kode" validate:"required,max=20"`

	Tipe      string   `json:"tipe" validate:"omitempty,oneof=pusat cabang"`
	Alamat    string   `json:"alamat" validate:"max=255"`
	PIC       string   `json:"pic" validate:"max=100"`
	Telepon   string   `json:"telepon" validate:"max=20"`
	Kapasitas int      `json:"kapasitas" validate:"min=0"`
	Latitude  *float64 `json:"latitude" validate:"omitempty,min=-90,max=90"`
	Longitude *float64 `json:"longitude" validate:"omitempty,min=-180,max=180"`
}

type ProtectRequest struct {
	IsProtected *bool `json:"is_protected" validate:"required"`
}
