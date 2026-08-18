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
	Username string `json:"username" validate:"required,min=4,max=50"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	FullName string `json:"full_name" validate:"required"`
	RoleID   uint   `json:"role_id" validate:"required"`
}

// UpdateUserRequest handles admin-side user edits (Manajemen User).
type UpdateUserRequest struct {
	Email    string `json:"email" validate:"omitempty,email"`
	FullName string `json:"full_name"`
	RoleID   uint   `json:"role_id"`
	IsActive *bool  `json:"is_active"`
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
