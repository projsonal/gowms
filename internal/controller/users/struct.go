package users

import (
	"time"

	"github.com/projsonal/gowms/internal/repositories/role"
	"github.com/projsonal/gowms/internal/repositories/users"
	"github.com/projsonal/gowms/pkg/config"
	"github.com/projsonal/gowms/pkg/otp"
	"github.com/projsonal/gowms/pkg/utils"
	"github.com/projsonal/gowms/pkg/wa"
)

// Controller menangani endpoint HTTP modul Manajemen User & Settings.
type Controller struct {
	userRepo users.Repository
	roleRepo role.Repository
	jwtSvc   *utils.JWTService
	waOTPSvc *otp.Service
	waSender wa.Sender
	waOTPTTL time.Duration
}

type Params struct {
	UserRepo users.Repository
	RoleRepo role.Repository
	JWTSvc   *utils.JWTService
	WaOTPSvc *otp.Service
	WaSender wa.Sender
	Cfg      *config.Config
}

// New membuat instance Controller Manajemen User.
func New(p Params) *Controller {
	return &Controller{
		userRepo: p.UserRepo,
		roleRepo: p.RoleRepo,
		jwtSvc:   p.JWTSvc,
		waOTPSvc: p.WaOTPSvc,
		waSender: p.WaSender,
		waOTPTTL: time.Duration(p.Cfg.WAOTP.TTLMinutes) * time.Minute,
	}
}

type CreateUserRequest struct {
	Username string `json:"username" validate:"required,min=4,max=50"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	FullName string `json:"full_name" validate:"required"`
	RoleID   uint   `json:"role_id" validate:"required"`
}

type UpdateUserRequest struct {
	Email    string `json:"email" validate:"omitempty,email"`
	FullName string `json:"full_name"`
	RoleID   uint   `json:"role_id"`
	IsActive *bool  `json:"is_active"`
}

type RequestChangePasswordOTPRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
}

type RequestChangePasswordOTPResponse struct {
	OTPToken  string `json:"otp_token"`
	ExpiresIn int    `json:"expires_in_seconds"`
}

type ConfirmChangePasswordRequest struct {
	OTPToken    string `json:"otp_token" validate:"required"`
	OTPCode     string `json:"otp_code" validate:"required,len=6"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

type Response struct {
	ID           uint   `json:"id"`
	Username     string `json:"username"`
	Email        string `json:"email"`
	FullName     string `json:"full_name"`
	RoleID       uint   `json:"role_id"`
	RoleName     string `json:"role_name"`
	IsActive     bool   `json:"is_active"`
	Is2FAEnabled bool   `json:"is_2fa_enabled"`
}
