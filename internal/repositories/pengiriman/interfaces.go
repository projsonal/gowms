// Package pengiriman mengakses tabel pengiriman & pengiriman_tracking_points
// pada modul "Pengiriman": pengiriman barang (pickup/dropoff) lengkap
// dengan pelacakan posisi kurir secara live (lat/long).
package pengiriman

import (
	"time"

	"github.com/projsonal/gostock/internal/model"
	"github.com/projsonal/gostock/pkg/utils"
)

type Filter struct {
	Status       string
	GudangAsalID uint
	Jenis        string
}

type Repository interface {
	List(p utils.PaginationParams, f Filter) ([]model.Pengiriman, int64, error)
	FindByID(id uint) (*model.Pengiriman, error)
	FindByNomor(nomor string) (*model.Pengiriman, error)
	Create(pg *model.Pengiriman) error
	Update(pg *model.Pengiriman) error
	Delete(id uint) error

	Jadwalkan(id uint, namaKurir, teleponKurir string, estimasiTiba *time.Time) (*model.Pengiriman, error)
	Mulai(id uint) (*model.Pengiriman, error)

	// RecordLocation menyimpan ping posisi GPS kurir: menambah baris baru
	// ke riwayat (PengirimanTrackingPoint) SEKALIGUS memperbarui
	// LastLat/LastLng/LastLocationAt pada header, dalam satu transaksi.
	RecordLocation(id uint, lat, lng float64, kecepatan *float64, recordedAt time.Time) (*model.Pengiriman, error)
	ListTrackingPoints(pengirimanID uint, limit int) ([]model.PengirimanTrackingPoint, error)

	Selesaikan(id uint, catatan string) (*model.Pengiriman, error)
	Batalkan(id uint) (*model.Pengiriman, error)

	CountByStatus(status string) (int64, error)
}
