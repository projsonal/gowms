package barang_serial

import (
	"github.com/projsonal/gowms/internal/model"
	barangRepo "github.com/projsonal/gowms/internal/repositories/barang"
	barangSerialRepo "github.com/projsonal/gowms/internal/repositories/barang_serial"
	"github.com/projsonal/gowms/internal/repositories/role"
	"github.com/projsonal/gowms/pkg/utils"
)

type Controller struct {
	repo       barangSerialRepo.Repository
	barangRepo barangRepo.Repository
	roleRepo   role.Repository
	jwtSvc     *utils.JWTService
}

func New(repo barangSerialRepo.Repository, barangRepo barangRepo.Repository, roleRepo role.Repository, jwtSvc *utils.JWTService) *Controller {
	return &Controller{repo: repo, barangRepo: barangRepo, roleRepo: roleRepo, jwtSvc: jwtSvc}
}

type UpdateStatusRequest struct {
	Status  string `json:"status" validate:"required,oneof=tersedia terpasang rusak"`
	Catatan string `json:"catatan" validate:"max=255"`
}

type CreateRequest struct {
	BarangID     uint   `json:"barang_id" validate:"required"`
	GudangID     uint   `json:"gudang_id" validate:"required"`
	SerialNumber string `json:"serial_number" validate:"required,max=100"`
	Catatan      string `json:"catatan" validate:"max=255"`
}

type RingkasanResponse struct {
	BarangID  uint  `json:"barang_id"`
	Tersedia  int64 `json:"tersedia"`
	Terpasang int64 `json:"terpasang"`
	Rusak     int64 `json:"rusak"`
}

type DetailResponse struct {
	*model.BarangSerial
	NomorBarangMasuk  string `json:"nomor_barang_masuk"`
	NomorBarangKeluar string `json:"nomor_barang_keluar"`
}
