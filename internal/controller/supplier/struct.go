package supplier

import (
	"github.com/projsonal/gowms/internal/model"
	"github.com/projsonal/gowms/internal/repositories/role"
	supplierRepo "github.com/projsonal/gowms/internal/repositories/supplier"
	"github.com/projsonal/gowms/pkg/utils"
)

type Controller struct {
	repo     supplierRepo.Repository
	roleRepo role.Repository
	jwtSvc   *utils.JWTService
}

func New(repo supplierRepo.Repository, roleRepo role.Repository, jwtSvc *utils.JWTService) *Controller {
	return &Controller{repo: repo, roleRepo: roleRepo, jwtSvc: jwtSvc}
}

type SupplierRequest struct {
	Kode    string  `json:"kode" validate:"required,max=30"`
	Nama    string  `json:"nama" validate:"required,max=150"`
	PIC     *string `json:"pic" validate:"max=100"`
	Telepon string  `json:"telepon" validate:"max=20"`

	KerjasamaKurir string  `json:"kerjasama_kurir" validate:"max=255"`
	Alamat         string  `json:"alamat" validate:"max=255"`
	NPWP           *string `json:"npwp" validate:"max=25"`
	Catatan        string  `json:"catatan" validate:"max=255"`
}

type UpdateStatusRequest struct {
	IsActive *bool `json:"is_active" validate:"required"`
}

type ProtectRequest struct {
	IsProtected *bool `json:"is_protected" validate:"required"`
}

type SummaryResponse struct {
	TotalSupplier int64 `json:"total_supplier"`
	SupplierAktif int64 `json:"supplier_aktif"`
}

type SupplierResponse struct {
	model.Supplier

	TotalOrder int64 `json:"total_order"`

	Rating float64 `json:"rating"`
}
