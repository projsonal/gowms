package model

import "time"

// Asset — modul Manajemen Aset Gudang (pengganti Task Management).
// Mencakup 6 jenis aset: tiang, odc, ont, odp, olt, transportasi.
//
// Dua skema penomoran berbeda dipakai (lihat pkg/constant JenisAset*):
//   - tiang/odc/ont/odp/olt: aset yang punya titik koordinat (untuk
//     tracking lokasi di peta) diberi LabelRSD dengan format
//     "{KodeGudang}-RSD-{nomor urut per gudang}", mis. "BBU-RSD-0001".
//     Nomor urut RESET per gudang (lihat repositories/asset NextRSDNumber).
//   - transportasi: tidak punya koordinat tetap (kendaraan berpindah),
//     jadi diberi KodeBA dengan format "BA-{nomor urut global}", mis.
//     "BA-0001" (lihat repositories/asset NextBANumber).
//
// Hanya salah satu dari LabelRSD/KodeBA yang terisi tergantung JenisAset.
type Asset struct {
	ID     uint   `json:"id" gorm:"primaryKey"`
	Nama   string `json:"nama" gorm:"size:150;not null"`
	// JenisAset: tiang | odc | ont | odp | olt | transportasi.
	JenisAset string `json:"jenis_aset" gorm:"size:20;not null;index"`

	GudangID uint    `json:"gudang_id" gorm:"not null;index"`
	Gudang   *Gudang `json:"gudang,omitempty" gorm:"foreignKey:GudangID"`

	// LabelRSD: label pelacakan untuk aset berkoordinat (tiang/odc/ont/
	// odp/olt). Kosong untuk transportasi.
	LabelRSD string `json:"label_rsd" gorm:"size:40;uniqueIndex"`
	// KodeBA: nomor barang aset untuk aset tanpa koordinat tetap
	// (transportasi). Kosong untuk jenis aset lain.
	KodeBA string `json:"kode_ba" gorm:"size:20;uniqueIndex"`

	// Latitude/Longitude: titik koordinat lokasi aset di lapangan — WAJIB
	// diisi untuk tiang/odc/ont/odp/olt (dipakai untuk tracking di peta),
	// nil untuk transportasi.
	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`

	// Status kondisi aset saat ini: aktif | rusak | nonaktif.
	Status      string `json:"status" gorm:"size:20;not null;default:'aktif';index"`
	Keterangan  string `json:"keterangan" gorm:"size:500"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Asset) TableName() string { return "assets" }

// JenisAsetPunyaKoordinat — true untuk jenis aset yang perlu titik lokasi
// & label RSD (tiang/odc/ont/odp/olt), false untuk transportasi (kode BA).
func JenisAsetPunyaKoordinat(jenisAset string) bool {
	return jenisAset != "transportasi"
}
