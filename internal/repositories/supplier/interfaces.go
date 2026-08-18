package supplier

import (
	"github.com/projsonal/gowms/internal/model"
	"github.com/projsonal/gowms/pkg/utils"
)

type Filter struct {
	OnlyActive bool
}

type Repository interface {
	List(p utils.PaginationParams, f Filter) ([]model.Supplier, int64, error)
	FindByID(id uint) (*model.Supplier, error)
	FindByKode(kode string) (*model.Supplier, error)
	Create(s *model.Supplier) error
	Update(s *model.Supplier) error
	Delete(id uint) error

	CountAll() (int64, error)
	CountActive() (int64, error)

	// InUse memeriksa apakah supplier ini masih direferensikan oleh
	// dokumen lain (Purchase Order / Barang Masuk) — keduanya punya kolom
	// supplier_id dengan foreign key ke tabel ini (lihat model.PO,
	// model.BarangMasuk). Kalau masih dipakai, DELETE akan selalu gagal
	// di level database (FK constraint violation) dengan pesan generik
	// yang membingungkan pengguna ("gagal menghapus supplier") — dicek
	// dulu di sini supaya controller bisa kasih pesan yang jelas.
	InUse(id uint) (bool, error)

	// KurirStats menghitung "hasil pelayanan" (service outcome) untuk
	// sekumpulan nama kurir mitra: totalOrder = jumlah pengiriman yang
	// sudah benar-benar diproses (bukan draft/dibatalkan) atas nama
	// kurir-kurir itu; terkirim = jumlah yang berhasil sampai tujuan.
	// Rating (0-5) dihitung di layer controller dari kedua angka ini.
	KurirStats(kurirNames []string) (totalOrder int64, terkirim int64, err error)
}
