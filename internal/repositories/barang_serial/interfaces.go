package barang_serial

import (
	"github.com/projsonal/gowms/internal/model"
	"github.com/projsonal/gowms/pkg/utils"
)

type Filter struct {
	BarangID uint
	GudangID uint
	Status   string // tersedia | terpasang | rusak, kosong = semua
}

type Repository interface {
	List(p utils.PaginationParams, f Filter) ([]model.BarangSerial, int64, error)
	FindByID(id uint) (*model.BarangSerial, error)
	// FindBySerial — pencarian utama fitur "beda-kan barang fisik lewat
	// SN": dipakai untuk cari/scan satu unit spesifik dan lihat lokasi +
	// statusnya saat ini, walau KodeBarang-nya sama dengan unit lain.
	FindBySerial(serial string) (*model.BarangSerial, error)
	CountByBarang(barangID uint) (tersedia int64, terpasang int64, rusak int64, err error)
	// Create — pendaftaran unit MANUAL, di luar alur dokumen Barang Masuk.
	// Dipakai khusus untuk mendigitalisasi stok fisik yang sudah ada di
	// gudang SEBELUM modul pelacakan SN ini dipakai (padanan langsung
	// dengan field BarangRequest.Stok di internal/controller/barang, yang
	// punya kegunaan serupa untuk barang non-serial). Menaikkan
	// Barang.Stok +1 sekaligus supaya tetap sinkron dengan unit barunya.
	Create(barangID, gudangID uint, rakID *uint, serialNumber, catatan string) (*model.BarangSerial, error)

	// RiwayatDokumen — nomor dokumen Barang Masuk/Keluar asal & tujuan
	// unit ini (kalau ada), dipakai menampilkan "riwayat" satu SN di UI
	// tanpa perlu tabel histori terpisah — cukup ikuti jejak
	// BarangMasukItemID/BarangKeluarItemID yang sudah tersimpan di baris
	// model.BarangSerial itu sendiri. String kosong berarti belum/tidak
	// ada (mis. unit didaftarkan manual lewat Create(), bukan dari
	// dokumen Barang Masuk, jadi NomorMasuk selalu kosong).
	RiwayatDokumen(s *model.BarangSerial) (nomorMasuk string, nomorKeluar string, err error)
	// UpdateStatusManual — tandai rusak/tersedia di luar alur dokumen
	// Barang Masuk/Keluar (mis. ditemukan cacat saat stock opname).
	UpdateStatusManual(id uint, status string, catatan string) (*model.BarangSerial, error)
	Delete(id uint) error
}
