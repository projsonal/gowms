package barang_masuk

import (
	"time"

	barangRepo "github.com/projsonal/gowms/internal/repositories/barang"
	bmRepo "github.com/projsonal/gowms/internal/repositories/barang_masuk"
	gudangRepo "github.com/projsonal/gowms/internal/repositories/gudang"
	notificationRepo "github.com/projsonal/gowms/internal/repositories/notification"
	"github.com/projsonal/gowms/internal/repositories/role"
	"github.com/projsonal/gowms/pkg/utils"
)

type Controller struct {
	repo       bmRepo.Repository
	barangRepo barangRepo.Repository
	gudangRepo gudangRepo.Repository
	roleRepo   role.Repository
	jwtSvc     *utils.JWTService
	notifRepo  notificationRepo.Repository
}

func New(repo bmRepo.Repository, barangRepo barangRepo.Repository, gudangRepo gudangRepo.Repository,
	roleRepo role.Repository, jwtSvc *utils.JWTService,
	notifRepo notificationRepo.Repository) *Controller {
	return &Controller{
		repo: repo, barangRepo: barangRepo, gudangRepo: gudangRepo,
		roleRepo: roleRepo, jwtSvc: jwtSvc,
		notifRepo: notifRepo,
	}
}

type ItemRequest struct {
	BarangID    uint  `json:"barang_id" validate:"required"`
	RakID       *uint `json:"rak_id"`
	Qty         int   `json:"qty" validate:"required,min=1"`
	HargaSatuan int64 `json:"harga_satuan" validate:"min=0"`
}

type BMRequest struct {
	GudangID uint `json:"gudang_id" validate:"required"`
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
	Catatan string        `json:"catatan" validate:"max=255"`
	Items   []ItemRequest `json:"items" validate:"required,min=1,dive"`
}

// parseTanggalHarian mem-parse tanggal "YYYY-MM-DD" (format bawaan
// <input type="date">) — dipakai semua modul yang formnya punya field
// tanggal (Barang Masuk/Keluar, Stock Opname) supaya konsisten, alih-alih
// tiap modul menulis parsing sendiri-sendiri.
func parseTanggalHarian(raw string) (time.Time, error) {
	return time.Parse("2006-01-02", raw)
}

// CompleteBMRequest — dikirim ke PATCH /barang-masuk/:id/selesai. Cukup
// body kosong ({} atau tanpa body sama sekali) untuk dokumen yang semua
// itemnya barang non-serial. Untuk item yang barangnya IsSerialized,
// wajib disertakan baris di sini dengan SerialNumbers sejumlah persis
// Qty item tersebut (lihat model.Barang.IsSerialized).
type CompleteBMRequest struct {
	Items []ItemSerialInput `json:"items"`
}

type ItemSerialInput struct {
	BarangMasukItemID uint     `json:"barang_masuk_item_id" validate:"required"`
	SerialNumbers     []string `json:"serial_numbers"`
}

type SummaryResponse struct {
	TotalDokumen int64 `json:"total_dokumen"`
	Draft        int64 `json:"draft"`
	Selesai      int64 `json:"selesai"`
}
