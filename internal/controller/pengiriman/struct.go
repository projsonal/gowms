package pengiriman

import (
	"time"

	barangKeluarRepo "github.com/projsonal/gowms/internal/repositories/barang_keluar"
	gudangRepo "github.com/projsonal/gowms/internal/repositories/gudang"
	notificationRepo "github.com/projsonal/gowms/internal/repositories/notification"
	pgRepo "github.com/projsonal/gowms/internal/repositories/pengiriman"
	"github.com/projsonal/gowms/internal/repositories/role"
	"github.com/projsonal/gowms/pkg/utils"
)

type Controller struct {
	repo             pgRepo.Repository
	gudangRepo       gudangRepo.Repository
	barangKeluarRepo barangKeluarRepo.Repository
	roleRepo         role.Repository
	jwtSvc           *utils.JWTService
	notifRepo        notificationRepo.Repository
}

func New(repo pgRepo.Repository, gudangRepo gudangRepo.Repository, barangKeluarRepo barangKeluarRepo.Repository,
	roleRepo role.Repository, jwtSvc *utils.JWTService, notifRepo notificationRepo.Repository) *Controller {
	return &Controller{
		repo: repo, gudangRepo: gudangRepo, barangKeluarRepo: barangKeluarRepo, roleRepo: roleRepo, jwtSvc: jwtSvc,
		notifRepo: notifRepo,
	}
}

// ---- DTO ----

type PengirimanRequest struct {
	BarangKeluarID   *uint    `json:"barang_keluar_id"`
	GudangAsalID     uint     `json:"gudang_asal_id" validate:"required"`
	JenisPengambilan string   `json:"jenis_pengambilan" validate:"required,oneof=pickup dropoff"`
	NamaPenerima     string   `json:"nama_penerima" validate:"required,max=150"`
	TeleponPenerima  string   `json:"telepon_penerima" validate:"max=20"`
	AlamatTujuan     string   `json:"alamat_tujuan" validate:"max=255"`
	DestLat          *float64 `json:"dest_lat" validate:"omitempty,min=-90,max=90"`
	DestLng          *float64 `json:"dest_lng" validate:"omitempty,min=-180,max=180"`
	// TanggalKirim: string "YYYY-MM-DD" — lihat catatan lengkap di
	// internal/controller/barang_masuk/struct.go BMRequest.Tanggal soal
	// kenapa ini WAJIB string, bukan time.Time langsung (form HTML
	// <input type="date"> tidak pernah kirim RFC3339 penuh).
	TanggalKirim string `json:"tanggal_kirim" validate:"required"`
	Catatan      string `json:"catatan" validate:"max=255"`
}

type JadwalkanRequest struct {
	NamaKurir    string `json:"nama_kurir" validate:"required,max=100"`
	TeleponKurir string `json:"telepon_kurir" validate:"max=20"`
	// EstimasiTiba: string "YYYY-MM-DD" opsional, sama alasannya seperti
	// TanggalKirim di atas.
	EstimasiTiba string `json:"estimasi_tiba"`
}

func parseTanggalHarian(raw string) (time.Time, error) {
	return time.Parse("2006-01-02", raw)
}

// LokasiRequest — ping posisi GPS dari perangkat/aplikasi kurir. RecordedAt
// opsional (default: waktu server menerima request) supaya kurir yang
// jamnya tidak sinkron tetap tercatat wajar secara berurutan.
// Catatan: Lat/Lng SENGAJA tidak diberi tag "required" — 0.0 adalah
// koordinat sah (khatulistiwa/garis bujur nol), jadi validasinya cukup
// lewat rentang min/max saja.
type LokasiRequest struct {
	Lat          float64    `json:"lat" validate:"min=-90,max=90"`
	Lng          float64    `json:"lng" validate:"min=-180,max=180"`
	KecepatanKmh *float64   `json:"kecepatan_kmh" validate:"omitempty,min=0"`
	RecordedAt   *time.Time `json:"recorded_at"`
}

type SelesaikanRequest struct {
	Catatan string `json:"catatan" validate:"max=255"`
}

// ProtectRequest — form aksi "Protect" di action bar tabel (khusus
// super_admin). Sama pola dengan Gudang/Barang/Supplier/PO.
type ProtectRequest struct {
	IsProtected *bool `json:"is_protected" validate:"required"`
}

type SummaryResponse struct {
	TotalPengiriman int64 `json:"total_pengiriman"`
	DalamPerjalanan int64 `json:"dalam_perjalanan"`
	Terkirim        int64 `json:"terkirim"`
}
