package notification

import (
	"github.com/projsonal/gowms/internal/model"
	"github.com/projsonal/gowms/pkg/utils"
)

// Row — satu notifikasi ditambah status "sudah dibaca" untuk user yang
// sedang login, bentuk gabungan yang dikembalikan List().
type Row struct {
	model.Notification
	IsRead bool `json:"is_read"`
}

type Repository interface {
	// Create — kirim notifikasi baru. Isi salah satu: UserID (personal)
	// atau RoleTarget (broadcast per role, "all" untuk semua role).
	Create(n *model.Notification) error
	// List — notifikasi milik userID: yang personal (user_id = userID)
	// ATAU broadcast (role_target cocok role user ini / "all"), diurutkan
	// terbaru dulu, dengan status IsRead per user ini.
	List(userID uint, userRole string, p utils.PaginationParams) ([]Row, int64, error)
	// UnreadCount — dipakai badge angka merah di ikon lonceng, TANPA perlu
	// ambil seluruh daftar (lebih ringan, dipoll berkala oleh frontend).
	UnreadCount(userID uint, userRole string) (int64, error)
	// MarkRead — tandai SATU notifikasi sudah dibaca oleh userID.
	MarkRead(notificationID, userID uint) error
	// MarkAllRead — tandai SEMUA notifikasi (personal + broadcast relevan)
	// milik userID sebagai sudah dibaca sekaligus (tombol "Tandai semua dibaca").
	MarkAllRead(userID uint, userRole string) error
	// Dismiss — hapus SATU notifikasi dari daftar milik userID SAJA (lihat
	// catatan panjang di model.NotificationDismissed kenapa ini bukan
	// DELETE baris Notification-nya langsung). Idempotent — dipanggil
	// ulang untuk notifikasi yang sudah di-dismiss tidak error.
	Dismiss(notificationID, userID uint) error
}
