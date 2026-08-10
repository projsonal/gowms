// Package model berisi seluruh entitas GORM (tabel database) yang dipakai
// lintas layer repository & controller.
package model

import "time"

type User struct {
	ID                  uint       `json:"id" gorm:"primaryKey"`
	Username            string     `json:"username" gorm:"uniqueIndex;size:50;not null"`
	// Email OPSIONAL — form registrasi tidak lagi mengumpulkan email
	// (identitas login memakai username). Tidak ada uniqueIndex/not null
	// lagi supaya banyak user boleh punya email kosong sekaligus tanpa
	// bentrok constraint unik (pola sama seperti PhoneNumber di bawah).
	Email               string     `json:"email" gorm:"size:100"`
	PasswordHash        string     `json:"-" gorm:"size:255;not null"`
	FullName            string     `json:"full_name" gorm:"size:100"`
	PhoneNumber         string     `json:"phone_number" gorm:"size:20"`
	// PhoneVerifiedAt diisi setelah user berhasil verifikasi OTP nomor HP
	// saat registrasi (lihat auth_controller.go: RequestRegisterOTP /
	// VerifyRegisterOTP). Nil berarti belum diverifikasi.
	PhoneVerifiedAt     *time.Time `json:"phone_verified_at" gorm:"column:phone_verified_at"`
	// AvatarURL: path relatif ke file foto profil ter-upload (lihat
	// PATCH /users/me & POST /users/me/avatar) — kosong berarti pakai
	// avatar inisial default di frontend.
	AvatarURL           string     `json:"avatar_url" gorm:"size:255"`
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
