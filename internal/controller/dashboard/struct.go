package dashboard

import (
	barangRepo "github.com/projsonal/gowms/internal/repositories/barang"
	barangKeluarRepo "github.com/projsonal/gowms/internal/repositories/barang_keluar"
	barangMasukRepo "github.com/projsonal/gowms/internal/repositories/barang_masuk"
	gudangRepo "github.com/projsonal/gowms/internal/repositories/gudang"
	pengirimanRepo "github.com/projsonal/gowms/internal/repositories/pengiriman"
	purchaseOrderRepo "github.com/projsonal/gowms/internal/repositories/po"
	"github.com/projsonal/gowms/internal/repositories/role"
	stockOpnameRepo "github.com/projsonal/gowms/internal/repositories/stockOpname"
	supplierRepo "github.com/projsonal/gowms/internal/repositories/supplier"
	"github.com/projsonal/gowms/pkg/utils"
	"gorm.io/gorm"
)

type Controller struct {
	barangRepo       barangRepo.Repository
	gudangRepo       gudangRepo.Repository
	supplierRepo     supplierRepo.Repository
	poRepo           purchaseOrderRepo.Repository
	barangMasukRepo  barangMasukRepo.Repository
	barangKeluarRepo barangKeluarRepo.Repository
	stockOpnameRepo  stockOpnameRepo.Repository
	pengirimanRepo   pengirimanRepo.Repository
	roleRepo         role.Repository
	jwtSvc           *utils.JWTService
	db               *gorm.DB
}

func New(barangRepo barangRepo.Repository, gudangRepo gudangRepo.Repository, supplierRepo supplierRepo.Repository,
	poRepo purchaseOrderRepo.Repository, barangMasukRepo barangMasukRepo.Repository, barangKeluarRepo barangKeluarRepo.Repository,
	stockOpnameRepo stockOpnameRepo.Repository, pengirimanRepo pengirimanRepo.Repository,
	roleRepo role.Repository, jwtSvc *utils.JWTService, db *gorm.DB) *Controller {
	return &Controller{
		barangRepo: barangRepo, gudangRepo: gudangRepo, supplierRepo: supplierRepo,
		poRepo: poRepo, barangMasukRepo: barangMasukRepo, barangKeluarRepo: barangKeluarRepo,
		stockOpnameRepo: stockOpnameRepo, pengirimanRepo: pengirimanRepo,
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

type SupplierSummary struct {
	TotalSupplier int64 `json:"total_supplier"`
	SupplierAktif int64 `json:"supplier_aktif"`
}

type PurchaseOrderSummary struct {
	TotalPO             int64 `json:"total_po"`
	MenungguPersetujuan int64 `json:"menunggu_persetujuan"`
	Disetujui           int64 `json:"disetujui"`
}

type DokumenSummary struct {
	Draft   int64 `json:"draft"`
	Selesai int64 `json:"selesai"`
}

type PengirimanSummary struct {
	DalamPerjalanan int64 `json:"dalam_perjalanan"`
	Terkirim        int64 `json:"terkirim"`
}

type DashboardResponse struct {
	KelolaBarang  KelolaBarangSummary  `json:"kelola_barang"`
	Gudang        GudangSummary        `json:"gudang"`
	Supplier      SupplierSummary      `json:"supplier"`
	PurchaseOrder PurchaseOrderSummary `json:"purchase_order"`
	BarangMasuk   DokumenSummary       `json:"barang_masuk"`
	BarangKeluar  DokumenSummary       `json:"barang_keluar"`
	StockOpname   DokumenSummary       `json:"stock_opname"`
	Pengiriman    PengirimanSummary    `json:"pengiriman"`
}
