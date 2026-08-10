package barang_masuk

import (
	"time"

	barangRepo "github.com/projsonal/gowms/internal/repositories/barang"
	bmRepo "github.com/projsonal/gowms/internal/repositories/barang_masuk"
	gudangRepo "github.com/projsonal/gowms/internal/repositories/gudang"
	poRepo "github.com/projsonal/gowms/internal/repositories/po"
	"github.com/projsonal/gowms/internal/repositories/role"
	supplierRepo "github.com/projsonal/gowms/internal/repositories/supplier"
	"github.com/projsonal/gowms/pkg/utils"
)

type Controller struct {
	repo         bmRepo.Repository
	barangRepo   barangRepo.Repository
	gudangRepo   gudangRepo.Repository
	poRepo       poRepo.Repository
	supplierRepo supplierRepo.Repository
	roleRepo     role.Repository
	jwtSvc       *utils.JWTService
}

func New(repo bmRepo.Repository, barangRepo barangRepo.Repository, gudangRepo gudangRepo.Repository,
	poRepo poRepo.Repository, supplierRepo supplierRepo.Repository, roleRepo role.Repository, jwtSvc *utils.JWTService) *Controller {
	return &Controller{
		repo: repo, barangRepo: barangRepo, gudangRepo: gudangRepo,
		poRepo: poRepo, supplierRepo: supplierRepo, roleRepo: roleRepo, jwtSvc: jwtSvc,
	}
}

type ItemRequest struct {
	BarangID    uint  `json:"barang_id" validate:"required"`
	RakID       *uint `json:"rak_id"`
	Qty         int   `json:"qty" validate:"required,min=1"`
	HargaSatuan int64 `json:"harga_satuan" validate:"min=0"`
}

type BMRequest struct {
	PurchaseOrderID *uint         `json:"purchase_order_id"`
	SupplierID      *uint         `json:"supplier_id"`
	GudangID        uint          `json:"gudang_id" validate:"required"`
	// Tanggal: SENGAJA string "YYYY-MM-DD" (bukan time.Time langsung) —
	// form HTML <input type="date"> di frontend cuma kirim tanggal polos
	// tanpa jam/zona waktu (mis. "2026-08-10"), sedangkan JSON unmarshal
	// bawaan Go untuk time.Time WAJIB format RFC3339 penuh
	// ("2026-08-10T00:00:00Z"). Kalau field ini langsung time.Time,
	// c.BodyParser() SELALU gagal untuk payload dari form ini dengan
	// pesan "payload tidak valid" — bukan soal izin/permission sama
	// sekali, murni ketidakcocokan format tanggal. Diparse manual di
	// Create()/Update() pakai parseTanggalHarian().
	Tanggal string        `json:"tanggal" validate:"required"`
	Catatan         string        `json:"catatan" validate:"max=255"`
	Items           []ItemRequest `json:"items" validate:"required,min=1,dive"`
}

// parseTanggalHarian mem-parse tanggal "YYYY-MM-DD" (format bawaan
// <input type="date">) — dipakai semua modul yang formnya punya field
// tanggal (Barang Masuk/Keluar, Pengiriman, Purchase Order, Stock Opname)
// supaya konsisten, alih-alih tiap modul menulis parsing sendiri-sendiri.
func parseTanggalHarian(raw string) (time.Time, error) {
	return time.Parse("2006-01-02", raw)
}

type SummaryResponse struct {
	TotalDokumen int64 `json:"total_dokumen"`
	Draft        int64 `json:"draft"`
	Selesai      int64 `json:"selesai"`
}
