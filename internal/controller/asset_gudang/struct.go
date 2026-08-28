package assetgudang

import (
	"time"

	assetRepo "github.com/projsonal/gowms/internal/repositories/asset"
	assetHistoryRepo "github.com/projsonal/gowms/internal/repositories/asset_history"
	assetPortRepo "github.com/projsonal/gowms/internal/repositories/asset_port"
	barangRepo "github.com/projsonal/gowms/internal/repositories/barang"
	gudangRepo "github.com/projsonal/gowms/internal/repositories/gudang"
	notificationRepo "github.com/projsonal/gowms/internal/repositories/notifikasi"
	"github.com/projsonal/gowms/internal/repositories/role"
	usersRepo "github.com/projsonal/gowms/internal/repositories/users"
	"github.com/projsonal/gowms/pkg/constant"
	"github.com/projsonal/gowms/pkg/utils"
)

const Module = constant.ModuleAsetGudang

type Controller struct {
	repo        assetRepo.Repository
	gudangRepo  gudangRepo.Repository
	portRepo    assetPortRepo.Repository
	historyRepo assetHistoryRepo.Repository
	usersRepo   usersRepo.Repository
	roleRepo    role.Repository
	barangRepo  barangRepo.Repository
	jwtSvc      *utils.JWTService
	notifRepo   notificationRepo.Repository
}

func New(repo assetRepo.Repository, gudangRepo gudangRepo.Repository, portRepo assetPortRepo.Repository, historyRepo assetHistoryRepo.Repository, usersRepo usersRepo.Repository, roleRepo role.Repository, barangRepo barangRepo.Repository, jwtSvc *utils.JWTService, notifRepo notificationRepo.Repository) *Controller {
	return &Controller{repo: repo, gudangRepo: gudangRepo, portRepo: portRepo, historyRepo: historyRepo, usersRepo: usersRepo, roleRepo: roleRepo, barangRepo: barangRepo, jwtSvc: jwtSvc, notifRepo: notifRepo}
}

type AssetRequest struct {
	Nama      string   `json:"nama" validate:"required,max=150"`
	JenisAset string   `json:"jenis_aset" validate:"required,oneof=tiang odc olt ont odp modem transportasi"`
	GudangID  uint     `json:"gudang_id" validate:"required"`
	Latitude  *float64 `json:"latitude" validate:"omitempty,min=-90,max=90"`
	Longitude *float64 `json:"longitude" validate:"omitempty,min=-180,max=180"`

	ParentAssetID *uint `json:"parent_asset_id"`

	JumlahPort int    `json:"jumlah_port" validate:"omitempty,min=0,max=512"`
	Merek      string `json:"merek" validate:"max=100"`
	Tipe       string `json:"tipe" validate:"max=100"`

	BarangID   *uint  `json:"barang_id"`
	Keterangan string `json:"keterangan" validate:"max=500"`
}

type UpdateStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=aktif rusak nonaktif"`
}

type SummaryResponse struct {
	Tiang        int64 `json:"tiang"`
	Odc          int64 `json:"odc"`
	Olt          int64 `json:"olt"`
	Ont          int64 `json:"ont"`
	Odp          int64 `json:"odp"`
	Modem        int64 `json:"modem"`
	Transportasi int64 `json:"transportasi"`
	Total        int64 `json:"total"`
}

type MapPoint struct {
	ID         uint    `json:"id"`
	Nama       string  `json:"nama"`
	JenisAset  string  `json:"jenis_aset"`
	LabelRSD   string  `json:"label_rsd"`
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
	Status     string  `json:"status"`
	GudangID   uint    `json:"gudang_id"`
	GudangNama string  `json:"gudang_nama"`
	GudangKode string  `json:"gudang_kode"`
	GudangTipe string  `json:"gudang_tipe"`

	GudangLatitude  *float64 `json:"gudang_latitude"`
	GudangLongitude *float64 `json:"gudang_longitude"`

	ParentAssetID   *uint    `json:"parent_asset_id"`
	ParentLatitude  *float64 `json:"parent_latitude"`
	ParentLongitude *float64 `json:"parent_longitude"`
	JumlahPort      int      `json:"jumlah_port"`
	PortTerisi      int64    `json:"port_terisi"`

	Merek      string `json:"merek,omitempty"`
	Tipe       string `json:"tipe,omitempty"`
	KodeBarang string `json:"kode_barang,omitempty"`
}

type AssetPortRequest struct {
	ChildAssetID  *uint  `json:"child_asset_id"`
	CustomerName  string `json:"customer_name" validate:"max=150"`
	CustomerPhone string `json:"customer_phone" validate:"max=20"`
	Keterangan    string `json:"keterangan" validate:"max=255"`
}

type AssetHistoryResponse struct {
	ID        uint      `json:"id"`
	EventType string    `json:"event_type"`
	FieldLama string    `json:"field_lama,omitempty"`
	FieldBaru string    `json:"field_baru,omitempty"`
	Catatan   string    `json:"catatan,omitempty"`
	UserNama  string    `json:"user_nama,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type AssetPortResponse struct {
	PortNumber      int    `json:"port_number"`
	Status          string `json:"status"`
	ChildAssetID    *uint  `json:"child_asset_id,omitempty"`
	ChildAssetNama  string `json:"child_asset_nama,omitempty"`
	ChildAssetLabel string `json:"child_asset_label,omitempty"`
	CustomerName    string `json:"customer_name,omitempty"`
	CustomerPhone   string `json:"customer_phone,omitempty"`
	Keterangan      string `json:"keterangan,omitempty"`
}
