package model

import "time"

// AssetHistory — riwayat perubahan SATU aset dari waktu ke waktu: inti
// dari fitur "tracking aset" yang sebenarnya (bukan cuma lihat status
// terkini, tapi tahu APA yang berubah, KAPAN, dan SIAPA yang mengubah).
// Satu baris = satu kejadian, TIDAK PERNAH diubah/dihapus setelah dibuat
// (append-only log) — kalau aset dihapus (soft-delete), riwayatnya tetap
// ada supaya jejak audit tidak hilang.
type AssetHistory struct {
	ID      uint `json:"id" gorm:"primaryKey"`
	AssetID uint `json:"asset_id" gorm:"not null;index"`

	// EventType: "dibuat" | "status" | "lokasi" | "induk" | "ping" | "port".
	// Menentukan bagaimana FieldLama/FieldBaru dibaca frontend (mis. kalau
	// EventType="status", FieldLama/FieldBaru berisi nilai status lama/baru).
	EventType string `json:"event_type" gorm:"size:20;not null;index"`
	FieldLama string `json:"field_lama" gorm:"size:255"`
	FieldBaru string `json:"field_baru" gorm:"size:255"`
	Catatan   string `json:"catatan" gorm:"size:255"`

	UserID   *uint  `json:"user_id"`
	UserNama string `json:"user_nama" gorm:"size:150"` // disalin saat kejadian, supaya riwayat tetap terbaca walau user-nya kemudian dihapus

	CreatedAt time.Time `json:"created_at"`
}

func (AssetHistory) TableName() string { return "asset_histories" }
