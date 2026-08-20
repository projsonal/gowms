package dashboard

import (
	barangRepo "github.com/projsonal/gowms/internal/repositories/barang"
	barangKeluarRepo "github.com/projsonal/gowms/internal/repositories/barang_keluar"
	barangMasukRepo "github.com/projsonal/gowms/internal/repositories/barang_masuk"
	gudangRepo "github.com/projsonal/gowms/internal/repositories/gudang"
	"github.com/projsonal/gowms/internal/repositories/role"
	stockOpnameRepo "github.com/projsonal/gowms/internal/repositories/stockOpname"
	"github.com/projsonal/gowms/pkg/utils"
	"gorm.io/gorm"
)

type Controller struct {
	barangRepo       barangRepo.Repository
	gudangRepo       gudangRepo.Repository
	barangMasukRepo  barangMasukRepo.Repository
	barangKeluarRepo barangKeluarRepo.Repository
	stockOpnameRepo  stockOpnameRepo.Repository
	roleRepo         role.Repository
	jwtSvc           *utils.JWTService
	db               *gorm.DB
}

func New(barangRepo barangRepo.Repository, gudangRepo gudangRepo.Repository,
	barangMasukRepo barangMasukRepo.Repository, barangKeluarRepo barangKeluarRepo.Repository,
	stockOpnameRepo stockOpnameRepo.Repository,
	roleRepo role.Repository, jwtSvc *utils.JWTService, db *gorm.DB) *Controller {
	return &Controller{
		barangRepo: barangRepo, gudangRepo: gudangRepo,
		barangMasukRepo: barangMasukRepo, barangKeluarRepo: barangKeluarRepo,
		stockOpnameRepo: stockOpnameRepo,
		roleRepo: roleRepo, jwtSvc: jwtSvc, db: db,
	}
}

type KelolaBarangSummary struct {
	TotalBarang          int64 `json:"total_barang"`
	StokMenipis          int64 `json:"stok_menipis"`
	TotalNilaiInventaris int64 `json:"total_nilai_inventaris"`
}

type GudangSummary struct {
	TotalGudang int64 `json:"total_gudang"`
	TotalRak    int64 `json:"total_rak"`
	RakPenuh    int64 `json:"rak_penuh"`
	RakKosong   int64 `json:"rak_kosong"`
}

type DokumenSummary struct {
	Draft   int64 `json:"draft"`
	Selesai int64 `json:"selesai"`
}

type DashboardResponse struct {
	KelolaBarang KelolaBarangSummary `json:"kelola_barang"`
	Gudang       GudangSummary       `json:"gudang"`
	BarangMasuk  DokumenSummary      `json:"barang_masuk"`
	BarangKeluar DokumenSummary      `json:"barang_keluar"`
	StockOpname  DokumenSummary      `json:"stock_opname"`
}
