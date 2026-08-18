package asset

import (
	"github.com/projsonal/gowms/internal/model"
	"github.com/projsonal/gowms/pkg/utils"
)

type Filter struct {
	JenisAset string
	GudangID  uint
	Status    string
}

// MapRow — hasil query gabungan assets+gudangs untuk Peta Sebaran Aset.
// Cuma kolom yang dipakai marker, supaya query & payload ringan saat
// mengambil seluruh titik tanpa paginasi (lihat Repository.ListForMap).
type MapRow struct {
	ID              uint
	Nama            string
	JenisAset       string
	LabelRSD        string
	Latitude        float64
	Longitude       float64
	Status          string
	IPAddress       string
	PingStatus      string
	GudangID        uint
	GudangNama      string
	GudangKode      string
	GudangTipe      string
	GudangLatitude  *float64
	GudangLongitude *float64
	ParentAssetID   *uint
	ParentLatitude  *float64
	ParentLongitude *float64
	JumlahPort      int
	PortTerisi      int64
}

type Repository interface {
	List(p utils.PaginationParams, f Filter) ([]model.Asset, int64, error)
	FindByID(id uint) (*model.Asset, error)
	// ListForMap — semua titik aset berkoordinat (lat/lng keduanya wajib
	// terisi) yang cocok dengan filter, TANPA paginasi. tipeGudang
	// ("pusat"/"cabang"/"") memfilter berdasar constant.TipeGudang* milik
	// gudang pemilik aset; string kosong berarti tidak difilter.
	ListForMap(f Filter, tipeGudang string) ([]MapRow, error)
	Create(a *model.Asset) error
	Update(a *model.Asset) error
	Delete(id uint) error

	// NextRSDNumber — nomor urut berikutnya untuk label RSD di gudang
	// tertentu (reset per gudang, dihitung dari total aset berkoordinat
	// yang sudah ada di gudang itu + 1).
	NextRSDNumber(gudangID uint) (int, error)
	// NextBANumber — nomor urut berikutnya untuk kode BA (global, lintas
	// gudang) khusus aset transportasi.
	NextBANumber() (int, error)

	CountByJenis(jenisAset string) (int64, error)
}
