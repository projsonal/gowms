// Package stock_opname mengakses tabel stock_opname & stock_opname_items
// pada modul "Stock Opname": pencocokan stok sistem vs hasil hitung fisik
// manual di lapangan. Tidak ada ketergantungan pada alat IoT/sensor —
// StokFisik murni diinput operator, dan selisihnya baru diterapkan ke stok
// Barang (& isi Rak bila diisi) saat dokumen diselesaikan (Complete).
package stock_opname

import (
	"github.com/projsonal/gostock/internal/model"
	"github.com/projsonal/gostock/pkg/utils"
)

type Filter struct {
	Status   string
	GudangID uint
}

// ItemInput dipakai saat menambah baris hitung ke opname draft — hanya
// perlu BarangID (+ RakID opsional), StokSistem diambil otomatis dari
// Barang.Stok saat itu juga supaya tidak bisa dimanipulasi dari client.
type ItemInput struct {
	BarangID  uint
	RakID     *uint
	StokFisik int
	Catatan   string
}

type Repository interface {
	List(p utils.PaginationParams, f Filter) ([]model.StockOpname, int64, error)
	FindByID(id uint) (*model.StockOpname, error)
	FindByNomor(nomor string) (*model.StockOpname, error)
	Create(so *model.StockOpname, inputs []ItemInput) error
	Update(so *model.StockOpname, inputs []ItemInput) error
	Delete(id uint) error

	// Complete menerapkan selisih (StokFisik - StokSistem) tiap item ke
	// Barang.Stok (dan Rak.Terisi bila item terkait rak tertentu), lalu
	// menandai dokumen "selesai" — semuanya dalam satu transaksi.
	Complete(id uint, userID uint) (*model.StockOpname, error)
	Batalkan(id uint) (*model.StockOpname, error)

	CountByStatus(status string) (int64, error)
}
