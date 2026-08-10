package purchase_order

import (
	"time"

	poRepo "github.com/projsonal/gowms/internal/repositories/po"
	"github.com/projsonal/gowms/internal/repositories/role"
	supplierRepo "github.com/projsonal/gowms/internal/repositories/supplier"
	"github.com/projsonal/gowms/pkg/utils"
)

type Controller struct {
	repo         poRepo.Repository
	supplierRepo supplierRepo.Repository
	roleRepo     role.Repository
	jwtSvc       *utils.JWTService
}

func New(repo poRepo.Repository, supplierRepo supplierRepo.Repository, roleRepo role.Repository, jwtSvc *utils.JWTService) *Controller {
	return &Controller{repo: repo, supplierRepo: supplierRepo, roleRepo: roleRepo, jwtSvc: jwtSvc}
}

// ---- DTO ----

type ItemRequest struct {
	BarangID    uint  `json:"barang_id" validate:"required"`
	QtyPesan    int   `json:"qty_pesan" validate:"required,min=1"`
	HargaSatuan int64 `json:"harga_satuan" validate:"min=0"`
}

type PORequest struct {
	SupplierID uint `json:"supplier_id" validate:"required"`
	// TanggalPO: string "YYYY-MM-DD" — lihat catatan lengkap di
	// internal/controller/barang_masuk/struct.go BMRequest.Tanggal.
	TanggalPO        string        `json:"tanggal_po" validate:"required"`
	CatatanPengajuan string        `json:"catatan_pengajuan" validate:"max=255"`
	Items            []ItemRequest `json:"items" validate:"required,min=1,dive"`
}

func parseTanggalHarian(raw string) (time.Time, error) {
	return time.Parse("2006-01-02", raw)
}

type SetujuiTolakRequest struct {
	Setuju  bool   `json:"setuju"`
	Catatan string `json:"catatan" validate:"max=255"`
}

type SummaryResponse struct {
	TotalPO             int64 `json:"total_po"`
	MenungguPersetujuan int64 `json:"menunggu_persetujuan"`
	Disetujui           int64 `json:"disetujui"`
	Selesai             int64 `json:"selesai"`
}
