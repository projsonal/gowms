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

// UpdateStatusRequest — form aksi manual tandai rusak/tersedia lagi di
// luar alur dokumen Barang Masuk/Keluar (mis. ditemukan cacat produksi
// saat stock opname). Lihat model.BarangSerial untuk arti tiap status.
type UpdateStatusRequest struct {
	Status  string `json:"status" validate:"required,oneof=tersedia terpasang rusak"`
	Catatan string `json:"catatan" validate:"max=255"`
}

// CreateRequest — pendaftaran unit MANUAL (di luar dokumen Barang Masuk),
// khusus untuk mendigitalisasi stok fisik yang sudah ada sebelum modul
// SN ini dipakai. Lihat barangSerialRepo.Repository.Create.
type CreateRequest struct {
	BarangID     uint   `json:"barang_id" validate:"required"`
	GudangID     uint   `json:"gudang_id" validate:"required"`
	RakID        *uint  `json:"rak_id"`
	SerialNumber string `json:"serial_number" validate:"required,max=100"`
	Catatan      string `json:"catatan" validate:"max=255"`
}

// RingkasanResponse — ringkasan jumlah unit per status untuk satu
// KodeBarang, ditampilkan di halaman detail barang (mis. "12 tersedia,
// 40 terpasang, 1 rusak").
type RingkasanResponse struct {
	BarangID  uint  `json:"barang_id"`
	Tersedia  int64 `json:"tersedia"`
	Terpasang int64 `json:"terpasang"`
	Rusak     int64 `json:"rusak"`
}

// DetailResponse — model.BarangSerial ditambah nomor dokumen Barang
// Masuk/Keluar asal & tujuannya (lihat
// barangSerialRepo.Repository.RiwayatDokumen) — inilah "riwayat" satu
// unit yang ditampilkan di UI, tanpa perlu tabel histori terpisah.
type DetailResponse struct {
	*model.BarangSerial
	NomorBarangMasuk  string `json:"nomor_barang_masuk"`
	NomorBarangKeluar string `json:"nomor_barang_keluar"`
}
