package model

import "time"

// Notification — notifikasi global di header (ikon lonceng), tampil di
// SEMUA menu untuk SEMUA role (bukan cuma menu Laporan). Dua mode target:
//   - UserID terisi -> ditujukan ke SATU user spesifik (mis. hasil
//     pengecekan barang rusak yang dia laporkan sendiri).
//   - UserID nil DAN RoleTarget terisi -> broadcast ke SEMUA user dengan
//     role tsb (mis. laporan barang rusak baru -> broadcast ke
//     super_admin & admin sekaligus, bukan tabel notifikasi terpisah per
//     penerima yang boros baris).
//
// IsRead per user disimpan terpisah lewat tabel NotificationRead (many-to-
// many) SUPAYA notifikasi broadcast tetap satu baris tapi status "sudah
// dibaca"-nya independen per user yang melihatnya.
type Notification struct {
	ID uint `json:"id" gorm:"primaryKey"`

	UserID     *uint  `json:"user_id" gorm:"index"`
	RoleTarget string `json:"role_target" gorm:"size:20;index"` // "" | "super_admin" | "admin" | "karyawan" | "all"

	// Type: dipakai frontend memilih ikon (mis. "barang_rusak", "ping",
	// "maintenance", "trash") — lihat NOTIF_TYPE_META di frontend.
	Type    string `json:"type" gorm:"size:30;not null"`
	Title   string `json:"title" gorm:"size:150;not null"`
	Message string `json:"message" gorm:"size:500"`
	// LinkHref: opsional, halaman tujuan saat notifikasi diklik (mis.
	// "/home/barang-rusak").
	LinkHref string `json:"link_href" gorm:"size:255"`

	CreatedAt time.Time `json:"created_at"`
}

func (Notification) TableName() string { return "notifications" }

// NotificationRead — penanda "user X sudah baca notifikasi Y", satu baris
// per (notification_id, user_id). Keberadaan baris = sudah dibaca.
type NotificationRead struct {
	NotificationID uint      `json:"notification_id" gorm:"primaryKey"`
	UserID         uint      `json:"user_id" gorm:"primaryKey"`
	ReadAt         time.Time `json:"read_at"`
}

func (NotificationRead) TableName() string { return "notification_reads" }

// NotificationDismissed — penanda "user X menghapus notifikasi Y DARI
// DAFTARNYA SENDIRI", satu baris per (notification_id, user_id). Notifikasi
// broadcast (RoleTarget terisi) dibagi SATU baris ke banyak user (lihat
// catatan di atas Notification), jadi "hapus" TIDAK BOLEH menghapus baris
// Notification itu sendiri — kalau dihapus, notifikasi ikut lenyap dari
// daftar semua user lain yang jadi target broadcast yang sama, bukan cuma
// dari daftar user yang menekan hapus. Pola ini persis sama seperti
// NotificationRead di atas: keberadaan baris = "disembunyikan buat user
// ini", baris Notification aslinya tetap utuh.
type NotificationDismissed struct {
	NotificationID uint      `json:"notification_id" gorm:"primaryKey"`
	UserID         uint      `json:"user_id" gorm:"primaryKey"`
	DismissedAt    time.Time `json:"dismissed_at"`
}

func (NotificationDismissed) TableName() string { return "notification_dismissed" }
