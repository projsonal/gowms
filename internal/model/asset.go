package model

import (
	"time"

	"gorm.io/gorm"
)

// Asset — modul Manajemen Aset Gudang & Peta Sebaran Aset (pengganti Task
// Management). Mencakup 7 jenis aset: tiang, odc, olt, ont, odp, modem,
// transportasi.
//
// Dua skema penomoran berbeda dipakai (lihat pkg/constant JenisAset*):
//   - tiang/odc/olt/ont/odp/modem: aset yang punya titik koordinat (untuk
//     tracking lokasi di peta, lihat internal/controller/asset_gudang
//     MapPoints & docs/peta-sebaran-aset.html) diberi LabelRSD dengan
//     format "{KodeGudang}-RSD-{nomor urut per gudang}", mis.
//     "BBU-RSD-0001" atau "MAHANG-RSD-0002". Nomor urut RESET per gudang
//     (lihat repositories/asset NextRSDNumber). KodeGudang berasal dari
//     model.Gudang.Kode, dan Gudang.Tipe (pusat/cabang) dipakai untuk
//     membedakan warna marker pusat vs cabang di peta.
//   - transportasi: tidak punya koordinat tetap (kendaraan berpindah),
//     jadi diberi KodeBA dengan format "BA-{nomor urut global}", mis.
//     "BA-0001" (lihat repositories/asset NextBANumber).
//
// Hanya salah satu dari LabelRSD/KodeBA yang terisi tergantung JenisAset.
type Asset struct {
	ID   uint   `json:"id" gorm:"primaryKey"`
	Nama string `json:"nama" gorm:"size:150;not null"`
	// JenisAset: tiang | odc | ont | odp | olt | transportasi.
	JenisAset string `json:"jenis_aset" gorm:"size:20;not null;index"`

	GudangID uint    `json:"gudang_id" gorm:"not null;index"`
	Gudang   *Gudang `json:"gudang,omitempty" gorm:"foreignKey:GudangID"`

	// LabelRSD: label pelacakan untuk aset berkoordinat (tiang/odc/ont/
	// odp/olt). Kosong untuk transportasi.
	//
	// SENGAJA TIDAK pakai tag `uniqueIndex` polos di sini: kolom ini kosong
	// ("") untuk SEMUA aset transportasi, dan di Postgres unique index
	// memperlakukan "" sebagai nilai biasa (beda dari NULL, yang boleh
	// berulang) — jadi aset transportasi KEDUA akan selalu dianggap
	// bentrok dengan yang pertama walau labelnya memang belum pernah
	// dipakai. Uniqueness untuk kolom ini ditegakkan lewat PARTIAL unique
	// index (`label_rsd <> ''`) yang dibuat manual di
	// pkg/config/database.go, bukan lewat tag gorm.
	LabelRSD string `json:"label_rsd" gorm:"size:40;index"`
	// KodeBA: nomor barang aset untuk aset tanpa koordinat tetap
	// (transportasi). Kosong untuk jenis aset lain.
	//
	// Alasan sama seperti LabelRSD di atas (dan sebelumnya inilah biang
	// kerok utama error "nomor label aset ini kebetulan sudah dipakai" saat
	// menambah aset OLT/tiang/dll: kolom ini kosong "" di SEMUA aset
	// berkoordinat, jadi aset non-transportasi KEDUA selalu bentrok di sini
	// duluan, walau pesan errornya menyebut "label"). Uniqueness lewat
	// partial index (`kode_ba <> ''`), bukan tag gorm.
	KodeBA string `json:"kode_ba" gorm:"size:20;index"`

	// Latitude/Longitude: titik koordinat lokasi aset di lapangan — WAJIB
	// diisi untuk tiang/odc/ont/odp/olt (dipakai untuk tracking di peta),
	// nil untuk transportasi.
	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`

	// ParentAssetID: hierarki jaringan FTTH — mis. ODP anak dari ODC, ODC
	// anak dari OLT (persis pola POP->OLT->ODC->ODP di referensi). Dipakai
	// menggambar garis kabel yang BENAR (aset ke aset induk jaringan),
	// bukan cuma aset ke gudang. Nil kalau aset ini titik teratas hierarki
	// (mis. OLT langsung di gudang) atau memang tidak relevan (tiang,
	// transportasi).
	ParentAssetID *uint  `json:"parent_asset_id" gorm:"index"`
	Parent        *Asset `json:"parent,omitempty" gorm:"foreignKey:ParentAssetID"`

	// JumlahPort: total slot port fisik perangkat (relevan untuk
	// odc/odp/olt — perangkat splitter/switch yang punya port ke
	// perangkat/pelanggan di bawahnya, lihat model.AssetPort). 0 berarti
	// aset ini tidak punya port (tiang, ont per-rumah, transportasi).
	JumlahPort int `json:"jumlah_port" gorm:"not null;default:0"`

	// Status kondisi aset saat ini: aktif | rusak | nonaktif.
	Status     string `json:"status" gorm:"size:20;not null;default:'aktif';index"`
	Keterangan string `json:"keterangan" gorm:"size:500"`

	// --- Tracking konektivitas via ping (khusus aset berkoordinat yang
	// punya alamat IP: odc/ont/odp/olt, kadang juga tiang kalau dipasangi
	// perangkat aktif) — lihat pkg/netping & Controller.Ping di
	// internal/controller/asset_gudang.
	//
	// IPAddress: alamat IP perangkat di lapangan, opsional. Kosong berarti
	// aset ini tidak dipantau via ping (mis. tiang polos tanpa perangkat).
	IPAddress string `json:"ip_address" gorm:"size:45"`
	// PingStatus: "online" | "offline" | "unknown" (default, belum pernah
	// dicek atau IPAddress belum diisi). TIDAK mengubah Status di atas
	// secara otomatis — murni indikator konektivitas terpisah dari status
	// kondisi fisik aset yang tetap dikontrol manual lewat UpdateStatus.
	PingStatus string     `json:"ping_status" gorm:"size:10;not null;default:'unknown'"`
	LastPingAt *time.Time `json:"last_ping_at"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	// DeletedAt: soft-delete GORM — lihat catatan lengkap di model.BarangRusak.
	// Dipulihkan/dihapus permanen lewat fitur Tempat Sampah (internal/controller/trash).
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (Asset) TableName() string { return "assets" }

// JenisAsetPunyaKoordinat — true untuk jenis aset yang perlu titik lokasi
// & label RSD (tiang/odc/ont/odp/olt), false untuk transportasi (kode BA).
func JenisAsetPunyaKoordinat(jenisAset string) bool {
	return jenisAset != "transportasi"
}

// jenisIndukValid — aturan longgar topologi FTTH standar: OLT di paling
// atas, lalu ODC, lalu ODP, lalu ONT/tiang di bawahnya. TIDAK memaksa
// urutan super ketat (mis. ODP boleh langsung ke OLT tanpa lewat ODC,
// desain jaringan nyata sering begitu) — cuma menolak kombinasi yang
// jelas terbalik/tidak masuk akal (mis. OLT jadi anak dari ODP).
var jenisIndukValid = map[string]map[string]bool{
	"odc":   {"olt": true},
	"odp":   {"olt": true, "odc": true},
	"ont":   {"olt": true, "odc": true, "odp": true},
	"tiang": {"olt": true, "odc": true, "odp": true, "tiang": true},
	"modem": {"olt": true, "odc": true, "odp": true, "ont": true},
}

// JenisIndukValid — true kalau `childJenis` boleh punya induk berjenis
// `parentJenis` secara topologi jaringan. Jenis tanpa aturan spesifik di
// atas (mis. "olt" sendiri, karena OLT biasanya paling atas) selalu
// dianggap valid (tidak dibatasi) — validasi ini SENGAJA longgar,
// bukan mesin aturan jaringan yang kaku.
func JenisIndukValid(childJenis, parentJenis string) bool {
	allowed, ok := jenisIndukValid[childJenis]
	if !ok {
		return true
	}
	return allowed[parentJenis]
}
