package assettype

import (
	assetTypeRepo "github.com/projsonal/gowms/internal/repositories/asset_type"
	"github.com/projsonal/gowms/pkg/constant"
	"github.com/projsonal/gowms/pkg/utils"
)

const Module = constant.ModuleAsetGudang

type Controller struct {
	repo   assetTypeRepo.Repository
	jwtSvc *utils.JWTService
}

func New(repo assetTypeRepo.Repository, jwtSvc *utils.JWTService) *Controller {
	return &Controller{repo: repo, jwtSvc: jwtSvc}
}

type AssetTypeRequest struct {
	Kode         string `json:"kode" validate:"required,max=30,alphanum"`
	Label        string `json:"label" validate:"required,max=60"`
	Color        string `json:"color" validate:"omitempty,max=10"`
	Abbr         string `json:"abbr" validate:"required,max=6"`
	HasKoordinat *bool  `json:"has_koordinat" validate:"required"`
	HasPort      *bool  `json:"has_port" validate:"required"`
	Urutan       int    `json:"urutan"`
}
