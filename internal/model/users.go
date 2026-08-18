// Package model berisi seluruh entitas GORM (tabel database) yang dipakai
// lintas layer repository & controller.
package model

import (
	"fmt"
	"time"
)

type User struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	Username string `json:"username" gorm:"uniqueIndex;size:50;not null"`
	// Email OPSIONAL — form registrasi tidak lagi mengumpulkan email
	// (identitas login memakai username). Tidak ada uniqueIndex/not null
	// lagi supaya banyak user boleh punya email kosong sekaligus tanpa
	// bentrok constraint unik (pola sama seperti PhoneNumber di bawah).
	Email        string `json:"email" gorm:"size:100"`
	PasswordHash string `json:"-" gorm:"size:255;not null"`
	FullName     string `json:"full_name" gorm:"size:100"`
	PhoneNumber  string `json:"phone_number" gorm:"size:20"`
	// PhoneVerifiedAt diisi setelah user berhasil verifikasi OTP nomor HP
	// saat registrasi (lihat auth_controller.go: RequestRegisterOTP /
	// VerifyRegisterOTP). Nil berarti belum diverifikasi.
	PhoneVerifiedAt *time.Time `json:"phone_verified_at" gorm:"column:phone_verified_at"`
	// AvatarData/AvatarContentType: isi foto profil disimpan LANGSUNG di
	// database (bytea), BUKAN di folder storage/ lokal — foto profil
	// termasuk data pribadi yang sensitif, jadi tidak boleh nongkrong
	// sebagai file statis yang bisa diakses siapa saja lewat URL publik
	// (lihat sebelumnya: app.Static("/uploads", ...) di router.go tidak
	// punya proteksi apa pun). Diisi lewat POST /users/me/avatar, dibaca
	// lewat GET /users/:id/avatar (WAJIB login — lihat user_controller.go
	// UploadAvatar/ServeAvatar). json:"-" karena binary besar tidak boleh
	// ikut ke-serialize ke payload JSON biasa (lihat Response.AvatarURL
	// yang dihitung dinamis, bukan field ini).
	AvatarData          []byte     `json:"-" gorm:"type:bytea"`
	AvatarContentType   string     `json:"-" gorm:"size:100"`
	RoleID              uint       `json:"role_id" gorm:"not null;index"`
	Role                *Role      `json:"role,omitempty" gorm:"foreignKey:RoleID"`
	IsActive            bool       `json:"is_active" gorm:"default:true"`
	Is2FAEnabled        bool       `json:"is_2fa_enabled" gorm:"column:is_2fa_enabled;default:false"`
	TOTPSecret          string     `json:"-" gorm:"column:totp_secret;size:100"`
	FailedLoginAttempts int        `json:"-" gorm:"default:0"`
	LockedUntil         *time.Time `json:"-"`
	LastLoginAt         *time.Time `json:"last_login_at"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

func (User) TableName() string { return "users" }

// AvatarURL menghitung URL untuk mengambil foto profil user ini lewat
// GET /users/:id/avatar (WAJIB login — lihat internal/controller/users
// user_controller.go ServeAvatar). Dipakai bareng oleh auth_controller.go
// (endpoint GET /auth/me) & user_controller.go (List/Detail/dst) supaya
// URL-nya dihitung dengan cara yang SAMA PERSIS di semua endpoint, tidak
// ada risiko dua tempat beda logika lalu salah satu ketinggalan saat
// avatar diubah. Kosong berarti user belum pernah upload foto — frontend
// jatuh ke avatar inisial. Query string ?v=<updated_at> memaksa
// browser/cache ambil ulang gambar setelah user ganti foto (nama URL-nya
// sendiri tidak berubah).
func (u User) AvatarURL() string {
	if len(u.AvatarData) == 0 {
		return ""
	}
	return fmt.Sprintf("/users/%d/avatar?v=%d", u.ID, u.UpdatedAt.Unix())
}
