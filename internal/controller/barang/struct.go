package barang

import (
	barangRepo "github.com/projsonal/gowms/internal/repositories/barang"
	gudangRepo "github.com/projsonal/gowms/internal/repositories/gudang"
	"github.com/projsonal/gowms/internal/repositories/role"
	"github.com/projsonal/gowms/pkg/utils"
)

type Controller struct {
	repo       barangRepo.Repository
	gudangRepo gudangRepo.Repository // dipakai memvalidasi KategoriID & SatuanID rujukan
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
	StokMinimum int    `json:"stok_minimum" validate:"min=0"`
	Deskripsi   string `json:"deskripsi" validate:"max=255"`
}

// AdjustStokRequest — dipakai modul lain (Barang Masuk/Keluar/Stock Opname)
// untuk menambah/mengurangi stok agregat barang tertentu.
type AdjustStokRequest struct {
	Delta int `json:"delta" validate:"required"`
}

// UpdateStatusRequest — form toggle aktif/nonaktif (didiskontinu) tanpa
// menghapus data barang, supaya histori transaksi lama tetap valid.
type UpdateStatusRequest struct {
	IsActive *bool `json:"is_active" validate:"required"`
}

// SummaryResponse — kartu "Total Barang | Stok Menipis | Total Nilai
// Inventaris" pada dashboard Kelola Barang.
type SummaryResponse struct {
	TotalBarang          int64 `json:"total_barang"`
	StokMenipis          int64 `json:"stok_menipis"`
	TotalNilaiInventaris int64 `json:"total_nilai_inventaris"`
}
