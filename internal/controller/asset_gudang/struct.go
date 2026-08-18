package assetgudang

import (
	"time"

	assetRepo "github.com/projsonal/gowms/internal/repositories/asset"
	assetHistoryRepo "github.com/projsonal/gowms/internal/repositories/asset_history"
	assetPortRepo "github.com/projsonal/gowms/internal/repositories/asset_port"
	gudangRepo "github.com/projsonal/gowms/internal/repositories/gudang"
	notificationRepo "github.com/projsonal/gowms/internal/repositories/notification"
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
	jwtSvc      *utils.JWTService
	notifRepo   notificationRepo.Repository
}

func New(repo assetRepo.Repository, gudangRepo gudangRepo.Repository, portRepo assetPortRepo.Repository, historyRepo assetHistoryRepo.Repository, usersRepo usersRepo.Repository, roleRepo role.Repository, jwtSvc *utils.JWTService, notifRepo notificationRepo.Repository) *Controller {
	return &Controller{repo: repo, gudangRepo: gudangRepo, portRepo: portRepo, historyRepo: historyRepo, usersRepo: usersRepo, roleRepo: roleRepo, jwtSvc: jwtSvc, notifRepo: notifRepo}
}

// AssetRequest — payload Tambah/Ubah Aset. Latitude & Longitude WAJIB
// diisi untuk jenis_aset selain "transportasi" (lihat Controller.Create).
type AssetRequest struct {
	Nama      string   `json:"nama" validate:"required,max=150"`
	JenisAset string   `json:"jenis_aset" validate:"required,oneof=tiang odc olt ont odp modem transportasi"`
	GudangID  uint     `json:"gudang_id" validate:"required"`
	Latitude  *float64 `json:"latitude" validate:"omitempty,min=-90,max=90"`
	Longitude *float64 `json:"longitude" validate:"omitempty,min=-180,max=180"`
	// IPAddress: alamat IP perangkat di lapangan, OPSIONAL — kalau diisi,
	// aset ini bisa dicek konektivitasnya lewat tombol "Cek Ping" (lihat
	// Controller.Ping). Kosong berarti tidak dipantau.
	IPAddress string `json:"ip_address" validate:"omitempty,ip"`
	// ParentAssetID: aset induk dalam hierarki jaringan FTTH (mis. ODP
	// ini anak dari ODC mana) — opsional, lihat model.Asset.ParentAssetID.
	ParentAssetID *uint `json:"parent_asset_id"`
	// JumlahPort: total slot port fisik perangkat ini (relevan untuk
	// odc/odp/olt) — opsional, default 0 (tidak punya port).
	JumlahPort int    `json:"jumlah_port" validate:"omitempty,min=0,max=512"`
	Keterangan string `json:"keterangan" validate:"max=500"`
}

// PingResponse — hasil satu kali pengecekan konektivitas aset.
type PingResponse struct {
	ID         uint       `json:"id"`
	IPAddress  string     `json:"ip_address"`
	PingStatus string     `json:"ping_status"`
	LastPingAt *time.Time `json:"last_ping_at"`
	RTTMs      int64      `json:"rtt_ms,omitempty"`
}

// UpdateStatusRequest — payload PATCH /aset/:id/status untuk menandai
// kondisi aset (mis. setelah pemeriksaan lapangan).
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

// MapPoint — bentuk ringkas Asset untuk Peta Sebaran Aset (GET /aset/map).
// Sengaja TIDAK memakai model.Asset penuh: endpoint ini dipanggil tanpa
// paginasi (bisa ratusan/ribuan titik sekaligus untuk dirender di Google
// Maps), jadi cuma field yang dipakai marker yang dikirim supaya payload-nya
// ringan.
type MapPoint struct {
	ID         uint    `json:"id"`
	Nama       string  `json:"nama"`
	JenisAset  string  `json:"jenis_aset"`
	LabelRSD   string  `json:"label_rsd"`
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
	Status     string  `json:"status"`
	IPAddress  string  `json:"ip_address"`
	PingStatus string  `json:"ping_status"`
	GudangID   uint    `json:"gudang_id"`
	GudangNama string  `json:"gudang_nama"`
	GudangKode string  `json:"gudang_kode"`
	GudangTipe string  `json:"gudang_tipe"` // "pusat" | "cabang"
	// GudangLatitude/GudangLongitude: dipakai frontend menggambar garis
	// "kabel" penghubung dari tiap titik aset ke gudang pemiliknya (lihat
	// halaman Tracking Aset) — nil kalau gudang itu belum diisi koordinat.
	GudangLatitude  *float64 `json:"gudang_latitude"`
	GudangLongitude *float64 `json:"gudang_longitude"`
	// ParentAssetID/ParentLatitude/ParentLongitude: kalau terisi, dipakai
	// frontend menggambar garis kabel hierarki JARINGAN (aset ke aset
	// induk) — INI YANG DIUTAMAKAN kalau ada, baru fallback ke garis ke
	// gudang kalau ParentAssetID kosong (lihat MapPoints handler).
	ParentAssetID   *uint    `json:"parent_asset_id"`
	ParentLatitude  *float64 `json:"parent_latitude"`
	ParentLongitude *float64 `json:"parent_longitude"`
	JumlahPort      int      `json:"jumlah_port"`
	PortTerisi      int64    `json:"port_terisi"`
}

// AssetPortRequest — payload isi/ubah satu port (PUT /aset/:id/port/:nomor).
type AssetPortRequest struct {
	// ChildAssetID XOR (CustomerName) — kalau ChildAssetID diisi, port ini
	// tersambung ke aset lain (hierarki); kalau tidak, dianggap tersambung
	// langsung ke pelanggan (CustomerName dkk).
	ChildAssetID  *uint  `json:"child_asset_id"`
	CustomerName  string `json:"customer_name" validate:"max=150"`
	CustomerPhone string `json:"customer_phone" validate:"max=20"`
	Keterangan    string `json:"keterangan" validate:"max=255"`
}

// AssetHistoryResponse — satu baris riwayat aset untuk timeline di frontend.
type AssetHistoryResponse struct {
	ID        uint      `json:"id"`
	EventType string    `json:"event_type"`
	FieldLama string    `json:"field_lama,omitempty"`
	FieldBaru string    `json:"field_baru,omitempty"`
	Catatan   string    `json:"catatan,omitempty"`
	UserNama  string    `json:"user_nama,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// AssetPortResponse — bentuk ringkas satu port untuk grid port di
// frontend (meniru grid "Port 1..8" di panel ODC/ODP referensi Fibero).
type AssetPortResponse struct {
	PortNumber      int    `json:"port_number"`
	Status          string `json:"status"` // "kosong" | "terisi"
	ChildAssetID    *uint  `json:"child_asset_id,omitempty"`
	ChildAssetNama  string `json:"child_asset_nama,omitempty"`
	ChildAssetLabel string `json:"child_asset_label,omitempty"`
	CustomerName    string `json:"customer_name,omitempty"`
	CustomerPhone   string `json:"customer_phone,omitempty"`
	Keterangan      string `json:"keterangan,omitempty"`
}
