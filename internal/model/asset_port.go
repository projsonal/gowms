package model

import (
	"time"

	"gorm.io/gorm"
)

// AssetPort — satu slot port fisik di perangkat jaringan (ODC/ODP/OLT,
// lihat Asset.JumlahPort), terinspirasi referensi Fibero (tiap ODP/ODC di
// sana menampilkan grid port berisi pelanggan yang terhubung).
//
// SATU port bisa terhubung ke SALAH SATU dari dua hal (tidak keduanya):
//   - ChildAssetID: port ini menyambung ke aset lain di bawahnya dalam
//     hierarki (mis. port ODC ke-3 tersambung ke ODP tertentu — lihat
//     juga Asset.ParentAssetID, dua arah relasi yang sama).
//   - Pelanggan (CustomerName dkk): port ini langsung ke rumah/pelanggan
//     (biasanya di level ODP, port terakhir sebelum ONT pelanggan).
//
// Port yang PortNumber-nya belum ada barisnya di tabel ini dianggap
// kosong/belum terpakai (Status "kosong") — tidak perlu insert baris
// untuk semua slot dari 1..JumlahPort di depan, cukup saat mulai dipakai.
type AssetPort struct {
	ID         uint `json:"id" gorm:"primaryKey"`
	AssetID    uint `json:"asset_id" gorm:"not null;index:idx_asset_port,unique"`
	PortNumber int  `json:"port_number" gorm:"not null;index:idx_asset_port,unique"`

	// Status: "kosong" | "terisi". Kolom terpisah dari sekadar "ada
	// barisnya atau tidak" supaya port yang PERNAH dipakai lalu
	// dikosongkan lagi (pelanggan berhenti) tetap tercatat riwayatnya
	// tanpa perlu dihapus barisnya.
	Status string `json:"status" gorm:"size:10;not null;default:'kosong'"`

	// ChildAssetID: kalau terisi, port ini tersambung ke aset lain (hierarki
	// jaringan) — lihat catatan lengkap di komentar struct di atas.
	ChildAssetID *uint  `json:"child_asset_id" gorm:"index"`
	ChildAsset   *Asset `json:"child_asset,omitempty" gorm:"foreignKey:ChildAssetID"`

	// Data pelanggan — SEDERHANA & OPSIONAL, sekadar catatan siapa yang
	// pakai port ini (bukan modul billing/CRM pelanggan penuh, itu di
	// luar cakupan WMS gudang ini). Kosong kalau port tersambung ke
	// ChildAssetID (aset lain), bukan pelanggan langsung.
	CustomerName  string `json:"customer_name" gorm:"size:150"`
	CustomerPhone string `json:"customer_phone" gorm:"size:20"`
	Keterangan    string `json:"keterangan" gorm:"size:255"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (AssetPort) TableName() string { return "asset_ports" }
