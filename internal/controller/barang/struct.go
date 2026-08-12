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
	// Stok — stok AWAL saat membuat SKU baru (mis. mendigitalisasi barang
	// yang sudah ada fisiknya di gudang sebelum sistem ini dipakai), atau
	// koreksi manual saat Ubah Barang (mis. hasil stok opname). Untuk
	// penambahan/pengurangan stok yang terikat transaksi (Barang Masuk/
	// Keluar), tetap pakai PATCH /:id/adjust supaya riwayatnya tercatat —
	// field ini untuk set NILAI ABSOLUT, bukan menambah/mengurangi.
	Stok        int    `json:"stok" validate:"min=0"`
	StokMinimum int    `json:"stok_minimum" validate:"min=0"`
	// BeratGram: opsional, lihat dokumentasi field di model/barang.go.
	BeratGram   *int   `json:"berat_gram" validate:"omitempty,min=0"`
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

// ProtectRequest — form aksi "Protect" di action bar tabel (khusus
// super_admin, lihat RegisterRoutes). Sama polanya dengan modul COD.
type ProtectRequest struct {
	IsProtected *bool `json:"is_protected" validate:"required"`
}

// RejectRequest — form aksi "Reject" saat super_admin menolak pengajuan
// barang dari admin (lihat model.Barang.ApprovalStatus).
type RejectRequest struct {
	Catatan string `json:"catatan" validate:"required,min=3"`
}

// SummaryResponse — kartu "Total Barang | Stok Menipis | Total Nilai
// Inventaris" pada dashboard Kelola Barang.
type SummaryResponse struct {
	TotalBarang          int64 `json:"total_barang"`
	StokMenipis          int64 `json:"stok_menipis"`
	TotalNilaiInventaris int64 `json:"total_nilai_inventaris"`
}
