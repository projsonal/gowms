package pengiriman

import (
	"time"

	barangKeluarRepo "github.com/projsonal/gostock/internal/repositories/barang_keluar"
	gudangRepo "github.com/projsonal/gostock/internal/repositories/gudang"
	pgRepo "github.com/projsonal/gostock/internal/repositories/pengiriman"
	"github.com/projsonal/gostock/internal/repositories/role"
	"github.com/projsonal/gostock/pkg/utils"
)

type Controller struct {
	repo             pgRepo.Repository
	gudangRepo       gudangRepo.Repository
	barangKeluarRepo barangKeluarRepo.Repository
	roleRepo         role.Repository
	jwtSvc           *utils.JWTService
}

func New(repo pgRepo.Repository, gudangRepo gudangRepo.Repository, barangKeluarRepo barangKeluarRepo.Repository,
	roleRepo role.Repository, jwtSvc *utils.JWTService) *Controller {
	return &Controller{
		repo: repo, gudangRepo: gudangRepo, barangKeluarRepo: barangKeluarRepo, roleRepo: roleRepo, jwtSvc: jwtSvc,
	}
}

// ---- DTO ----

type PengirimanRequest struct {
	BarangKeluarID   *uint     `json:"barang_keluar_id"`
	GudangAsalID     uint      `json:"gudang_asal_id" validate:"required"`
	JenisPengambilan string    `json:"jenis_pengambilan" validate:"required,oneof=pickup dropoff"`
	NamaPenerima     string    `json:"nama_penerima" validate:"required,max=150"`
	TeleponPenerima  string    `json:"telepon_penerima" validate:"max=20"`
	AlamatTujuan     string    `json:"alamat_tujuan" validate:"max=255"`
	TanggalKirim     time.Time `json:"tanggal_kirim" validate:"required"`
	Catatan          string    `json:"catatan" validate:"max=255"`
}

type JadwalkanRequest struct {
	NamaKurir    string     `json:"nama_kurir" validate:"required,max=100"`
	TeleponKurir string     `json:"telepon_kurir" validate:"max=20"`
	EstimasiTiba *time.Time `json:"estimasi_tiba"`
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

type SummaryResponse struct {
	TotalPengiriman int64 `json:"total_pengiriman"`
	DalamPerjalanan int64 `json:"dalam_perjalanan"`
	Terkirim        int64 `json:"terkirim"`
}
