package barang

import (
	barangRepo "github.com/projsonal/gowms/internal/repositories/barang"
	gudangRepo "github.com/projsonal/gowms/internal/repositories/gudang"
	"github.com/projsonal/gowms/internal/repositories/role"
	"github.com/projsonal/gowms/pkg/utils"
)

type Controller struct {
	repo       barangRepo.Repository
	gudangRepo gudangRepo.Repository
	roleRepo   role.Repository
	jwtSvc     *utils.JWTService
}

func New(repo barangRepo.Repository, gudangRepo gudangRepo.Repository, roleRepo role.Repository, jwtSvc *utils.JWTService) *Controller {
	return &Controller{repo: repo, gudangRepo: gudangRepo, roleRepo: roleRepo, jwtSvc: jwtSvc}
}

type BarangRequest struct {
	KodeBarang  string `json:"kode_barang" validate:"required,max=30"`
	Nama        string `json:"nama" validate:"required,max=150"`
	KategoriID  uint   `json:"kategori_id" validate:"required"`
	SatuanID    uint   `json:"satuan_id" validate:"required"`
	HargaBeli   int64  `json:"harga_beli" validate:"min=0"`
	StokAwal    int    `json:"stok" validate:"min=0"`
	StokMinimum int    `json:"stok_minimum" validate:"min=0"`
	BeratGram   *int   `json:"berat_gram" validate:"omitempty,min=0"`
	Deskripsi   string `json:"deskripsi" validate:"max=255"`
}

type AdjustStokRequest struct {
	Delta int `json:"delta" validate:"required"`
}

type UpdateStatusRequest struct {
	IsActive *bool `json:"is_active" validate:"required"`
}

type ProtectRequest struct {
	IsProtected *bool `json:"is_protected" validate:"required"`
}

type RejectRequest struct {
	Catatan string `json:"catatan" validate:"required,min=3"`
}

type SummaryResponse struct {
	TotalBarang          int64 `json:"total_barang"`
	StokMenipis          int64 `json:"stok_menipis"`
	TotalNilaiInventaris int64 `json:"total_nilai_inventaris"`
}
