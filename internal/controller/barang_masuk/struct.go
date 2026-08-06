package barang_masuk

import (
	"time"

	barangRepo "github.com/projsonal/gostock/internal/repositories/barang"
	bmRepo "github.com/projsonal/gostock/internal/repositories/barang_masuk"
	gudangRepo "github.com/projsonal/gostock/internal/repositories/gudang"
	poRepo "github.com/projsonal/gostock/internal/repositories/po"
	"github.com/projsonal/gostock/internal/repositories/role"
	supplierRepo "github.com/projsonal/gostock/internal/repositories/supplier"
	"github.com/projsonal/gostock/pkg/utils"
)

type Controller struct {
	repo         bmRepo.Repository
	barangRepo   barangRepo.Repository
	gudangRepo   gudangRepo.Repository
	poRepo       poRepo.Repository
	supplierRepo supplierRepo.Repository
	roleRepo     role.Repository
	jwtSvc       *utils.JWTService
}

func New(repo bmRepo.Repository, barangRepo barangRepo.Repository, gudangRepo gudangRepo.Repository,
	poRepo poRepo.Repository, supplierRepo supplierRepo.Repository, roleRepo role.Repository, jwtSvc *utils.JWTService) *Controller {
	return &Controller{
		repo: repo, barangRepo: barangRepo, gudangRepo: gudangRepo,
		poRepo: poRepo, supplierRepo: supplierRepo, roleRepo: roleRepo, jwtSvc: jwtSvc,
	}
}

type ItemRequest struct {
	BarangID    uint  `json:"barang_id" validate:"required"`
	RakID       *uint `json:"rak_id"`
	Qty         int   `json:"qty" validate:"required,min=1"`
	HargaSatuan int64 `json:"harga_satuan" validate:"min=0"`
}

type BMRequest struct {
	PurchaseOrderID *uint         `json:"purchase_order_id"`
	SupplierID      *uint         `json:"supplier_id"`
	GudangID        uint          `json:"gudang_id" validate:"required"`
	Tanggal         time.Time     `json:"tanggal" validate:"required"`
	Catatan         string        `json:"catatan" validate:"max=255"`
	Items           []ItemRequest `json:"items" validate:"required,min=1,dive"`
}

type SummaryResponse struct {
	TotalDokumen int64 `json:"total_dokumen"`
	Draft        int64 `json:"draft"`
	Selesai      int64 `json:"selesai"`
}
