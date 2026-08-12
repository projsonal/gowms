package supplier

import (
	"github.com/projsonal/gowms/internal/model"
	"github.com/projsonal/gowms/internal/repositories/role"
	supplierRepo "github.com/projsonal/gowms/internal/repositories/supplier"
	"github.com/projsonal/gowms/pkg/utils"
)

type Controller struct {
	repo     supplierRepo.Repository
	roleRepo role.Repository
	jwtSvc   *utils.JWTService
}

func New(repo supplierRepo.Repository, roleRepo role.Repository, jwtSvc *utils.JWTService) *Controller {
	return &Controller{repo: repo, roleRepo: roleRepo, jwtSvc: jwtSvc}
}

type SupplierRequest struct {
	Kode    string  `json:"kode" validate:"required,max=30"`
	Nama    string  `json:"nama" validate:"required,max=150"`
	PIC     *string `json:"pic" validate:"max=100"`
	Telepon string  `json:"telepon" validate:"max=20"`
	// KerjasamaKurir: daftar nama kurir yang bekerja sama dengan supplier
	// ini untuk pengiriman ke lokasi tujuan (dipisah koma, mis. "JNE,J&T,
	// Lalamove") — dipakai mencocokkan ke Pengiriman.NamaKurir untuk
	// menghitung TotalOrder & Rating (lihat SupplierWithStats).
	KerjasamaKurir string  `json:"kerjasama_kurir" validate:"max=255"`
	Alamat         string  `json:"alamat" validate:"max=255"`
	NPWP           *string `json:"npwp" validate:"max=25"`
	Catatan        string  `json:"catatan" validate:"max=255"`
}

type UpdateStatusRequest struct {
	IsActive *bool `json:"is_active" validate:"required"`
}

// ProtectRequest — form aksi "Protect" di action bar tabel (khusus
// super_admin, lihat RegisterRoutes). Sama polanya dengan modul COD/Barang.
type ProtectRequest struct {
	IsProtected *bool `json:"is_protected" validate:"required"`
}

type SummaryResponse struct {
	TotalSupplier int64 `json:"total_supplier"`
	SupplierAktif int64 `json:"supplier_aktif"`
}

// SupplierResponse membungkus model.Supplier dengan TotalOrder & Rating
// hasil hitung otomatis (bukan kolom tersimpan) — lihat withStats().
type SupplierResponse struct {
	model.Supplier
	// TotalOrder: jumlah pengiriman (bukan draft/dibatalkan) yang memakai
	// salah satu kurir di KerjasamaKurir supplier ini.
	TotalOrder int64 `json:"total_order"`
	// Rating: hasil pelayanan (0-5) = proporsi pengiriman yang berhasil
	// sampai tujuan ("terkirim") dari TotalOrder, diskalakan ke 5. 0 kalau
	// belum ada order sama sekali (bukan berarti performa buruk).
	Rating float64 `json:"rating"`
}
