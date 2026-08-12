package model

import "time"

type PurchaseOrder struct {
	ID               uint                `json:"id" gorm:"primaryKey"`
	NomorPO          string              `json:"nomor_po" gorm:"size:30;uniqueIndex;not null"` // PO-2026-0001
	SupplierID       uint                `json:"supplier_id" gorm:"not null;index"`
	Supplier         *Supplier           `json:"supplier,omitempty" gorm:"foreignKey:SupplierID"`
	Status           string              `json:"status" gorm:"size:20;not null;default:'draft';index"`
	TanggalPO        time.Time           `json:"tanggal_po" gorm:"not null"`
	CatatanPengajuan string              `json:"catatan_pengajuan" gorm:"size:255"`
	CatatanApproval  string              `json:"catatan_approval" gorm:"size:255"`
	DiajukanOleh     *uint               `json:"diajukan_oleh"`
	DiajukanAt       *time.Time          `json:"diajukan_at"`
	DisetujuiOleh    *uint               `json:"disetujui_oleh"`
	DisetujuiAt      *time.Time          `json:"disetujui_at"`
	TotalEstimasi    int64               `json:"total_estimasi" gorm:"not null;default:0"`
	IsProtected      bool                `json:"is_protected" gorm:"not null;default:false"` // dikunci super_admin
	Items            []PurchaseOrderItem `json:"items,omitempty" gorm:"foreignKey:PurchaseOrderID"`
	CreatedAt        time.Time           `json:"created_at"`
	UpdatedAt        time.Time           `json:"updated_at"`
}

func (PurchaseOrder) TableName() string { return "purchase_orders" }

// IsFullyReceived mengecek apakah seluruh item PO sudah diterima penuh
// (QtyDiterima >= QtyPesan untuk semua baris) — dipakai repository untuk
// menentukan kapan PO otomatis pindah status ke "selesai".
func (po *PurchaseOrder) IsFullyReceived() bool {
	if len(po.Items) == 0 {
		return false
	}
	for _, it := range po.Items {
		if it.QtyDiterima < it.QtyPesan {
			return false
		}
	}
	return true
}

type PurchaseOrderItem struct {
	ID              uint    `json:"id" gorm:"primaryKey"`
	PurchaseOrderID uint    `json:"purchase_order_id" gorm:"not null;index"`
	BarangID        uint    `json:"barang_id" gorm:"not null;index"`
	Barang          *Barang `json:"barang,omitempty" gorm:"foreignKey:BarangID"`
	QtyPesan        int     `json:"qty_pesan" gorm:"not null"`
	QtyDiterima     int     `json:"qty_diterima" gorm:"not null;default:0"`
	HargaSatuan     int64   `json:"harga_satuan" gorm:"not null;default:0"`
	Subtotal        int64   `json:"subtotal" gorm:"not null;default:0"`
}

func (PurchaseOrderItem) TableName() string { return "purchase_order_items" }

// SisaDiterima menghitung berapa unit yang masih perlu diterima dari item
// PO ini — dipakai Barang Masuk untuk membatasi qty yang boleh direalisasi
// supaya tidak menerima melebihi yang dipesan.
func (i *PurchaseOrderItem) SisaDiterima() int {
	sisa := i.QtyPesan - i.QtyDiterima
	if sisa < 0 {
		return 0
	}
	return sisa
}
