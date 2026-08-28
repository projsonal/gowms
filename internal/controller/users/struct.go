package users

import (
	"time"

	authRepo "github.com/projsonal/gowms/internal/repositories/auth"
	"github.com/projsonal/gowms/internal/repositories/role"
	"github.com/projsonal/gowms/internal/repositories/users"
	"github.com/projsonal/gowms/pkg/humancheck"
	"github.com/projsonal/gowms/pkg/utils"
)

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

type UpdateUserRequest struct {
	Email    string `json:"email" validate:"omitempty,email"`
	FullName string `json:"full_name"`

	PhoneNumber *string `json:"phone_number" validate:"omitempty"`
	RoleID      uint    `json:"role_id"`
	IsActive    *bool   `json:"is_active"`
}

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

	IsOnline     bool       `json:"is_online"`
	Is2FAEnabled bool       `json:"is_2fa_enabled"`
	LastLoginAt  *time.Time `json:"last_login_at"`
}

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
