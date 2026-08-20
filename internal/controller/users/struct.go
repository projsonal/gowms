package users

import (
	"time"

	authRepo "github.com/projsonal/gowms/internal/repositories/auth"
	"github.com/projsonal/gowms/internal/repositories/role"
	"github.com/projsonal/gowms/internal/repositories/users"
	"github.com/projsonal/gowms/pkg/humancheck"
	"github.com/projsonal/gowms/pkg/utils"
)

// Controller menangani endpoint HTTP modul Manajemen User & Settings.
type Controller struct {
	userRepo      users.Repository
	roleRepo      role.Repository
	authRepo      authRepo.Repository
	jwtSvc        *utils.JWTService
	humanCheckSvc *humancheck.Service
	storagePath   string
}

type Params struct {
	UserRepo      users.Repository
	RoleRepo      role.Repository
	AuthRepo      authRepo.Repository
	JWTSvc        *utils.JWTService
	HumanCheckSvc *humancheck.Service
	StoragePath   string
}

// New membuat instance Controller Manajemen User.
func New(p Params) *Controller {
	return &Controller{
		userRepo:      p.UserRepo,
		roleRepo:      p.RoleRepo,
		authRepo:      p.AuthRepo,
		jwtSvc:        p.JWTSvc,
		humanCheckSvc: p.HumanCheckSvc,
		storagePath:   p.StoragePath,
	}
}

type CreateUserRequest struct {
	Username    string `json:"username" validate:"required,min=4,max=50"`
	Email       string `json:"email" validate:"required,email"`
	Password    string `json:"password" validate:"required,min=8"`
	FullName    string `json:"full_name" validate:"required"`
	PhoneNumber string `json:"phone_number" validate:"omitempty,max=20"`
	RoleID      uint   `json:"role_id" validate:"required"`
}

// UpdateUserRequest handles admin-side user edits (Manajemen User).
type UpdateUserRequest struct {
	Email    string `json:"email" validate:"omitempty,email"`
	FullName string `json:"full_name"`
	// PhoneNumber pakai *string (bukan string polos) supaya bisa
	// dibedakan "field tidak dikirim sama sekali" (nil, jangan diubah)
	// dari "sengaja dikosongkan" (pointer ke string kosong, hapus nomor
	// yang tersimpan) — beda dari Email/FullName di atas yang memang
	// tidak punya kasus "sengaja dikosongkan lewat form Manajemen User".
	PhoneNumber *string `json:"phone_number" validate:"omitempty"`
	RoleID      uint    `json:"role_id"`
	IsActive    *bool   `json:"is_active"`
}

// ChangePasswordRequest — ganti password langsung dalam SATU langkah (tanpa
// OTP WhatsApp), diverifikasi lewat checkbox "verify you are human" ala
// Cloudflare Turnstile (lihat pkg/humancheck) supaya tetap ada perlindungan
// dari automated abuse tanpa menyuruh user memecahkan captcha gambar.
type ChangePasswordRequest struct {
	OldPassword     string `json:"old_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8"`
	HumanCheckToken string `json:"human_check_token" validate:"required"`
}

type Response struct {
	ID          uint   `json:"id"`
	Username    string `json:"username"`
	Email       string `json:"email"`
	FullName    string `json:"full_name"`
	PhoneNumber string `json:"phone_number"`
	AvatarURL   string `json:"avatar_url"`
	RoleID      uint   `json:"role_id"`
	RoleName    string `json:"role_name"`
	IsActive    bool   `json:"is_active"`
	// IsOnline: status login SAAT INI (punya sesi refresh token yang
	// belum dicabut & belum kedaluwarsa) — INI yang ditampilkan kolom
	// "Status" (Aktif/Nonaktif) di tabel Manajemen User, BUKAN IsActive
	// (itu flag akun diaktifkan/dinonaktifkan admin, konsep berbeda: akun
	// bisa "aktif" tapi user-nya sedang tidak login sama sekali).
	IsOnline     bool       `json:"is_online"`
	Is2FAEnabled bool       `json:"is_2fa_enabled"`
	LastLoginAt  *time.Time `json:"last_login_at"`
}

// SessionResponse — satu sesi login aktif milik user tertentu (perangkat
// mana saja user tsb sedang login), ditampilkan lewat aksi "Lihat
// Perangkat" di Manajemen User (khusus Super Admin — lihat RegisterRoutes
// UserSessions/RevokeUserSession). Padanan admin-side dari
// GET/DELETE /auth/sessions milik auth_controller.go yang cuma bisa lihat
// sesi DIRI SENDIRI; endpoint ini membiarkan Super Admin melihat & mencabut
// sesi user LAIN — makanya sengaja dipisah, bukan reuse endpoint yang sama.
type SessionResponse struct {
	ID             uint   `json:"id"`
	Browser        string `json:"browser"`
	BrowserVersion string `json:"browser_version"`
	OS             string `json:"os"`
	OSVersion      string `json:"os_version"`
	DeviceType     string `json:"device_type"`
	IPAddress      string `json:"ip_address"`
	Location       string `json:"location"`
	CreatedAt      string `json:"created_at"`
}

type SessionListResponse struct {
	Sessions []SessionResponse `json:"sessions"`
}
