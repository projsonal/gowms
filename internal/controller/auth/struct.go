package auth

import (
	"time"

	authRepo "github.com/projsonal/gowms/internal/repositories/auth"
	"github.com/projsonal/gowms/internal/repositories/role"
	"github.com/projsonal/gowms/internal/repositories/users"
	"github.com/projsonal/gowms/pkg/captcha"
	"github.com/projsonal/gowms/pkg/config"
	"github.com/projsonal/gowms/pkg/geoip"
	"github.com/projsonal/gowms/pkg/humancheck"
	"github.com/projsonal/gowms/pkg/utils"
)

const otpReplayTTL = 5 * time.Minute

type Controller struct {
	userRepo      users.Repository
	roleRepo      role.Repository
	authRepo      authRepo.Repository
	jwtSvc        *utils.JWTService
	captchaSvc    *captcha.Service
	humanCheckSvc *humancheck.Service
	geoipSvc      geoip.Resolver

	appEnv     string
	totpIssuer string

	otpReplayGuard *totpReplayGuard
}

type Params struct {
	UserRepo      users.Repository
	RoleRepo      role.Repository
	AuthRepo      authRepo.Repository
	JWTSvc        *utils.JWTService
	CaptchaSvc    *captcha.Service
	HumanCheckSvc *humancheck.Service
	GeoipSvc      geoip.Resolver
	Cfg           *config.Config
}

func New(p Params) *Controller {
	return &Controller{
		userRepo:      p.UserRepo,
		roleRepo:      p.RoleRepo,
		authRepo:      p.AuthRepo,
		jwtSvc:        p.JWTSvc,
		captchaSvc:    p.CaptchaSvc,
		humanCheckSvc: p.HumanCheckSvc,
		geoipSvc:      p.GeoipSvc,
		appEnv:        p.Cfg.App.Env,
		totpIssuer:    p.Cfg.TOTP.Issuer,

		otpReplayGuard: newTOTPReplayGuard(otpReplayTTL),
	}
}

type RegisterRequest struct {
	Username             string `json:"username" validate:"required,min=4,max=50"`
	Email                string `json:"email" validate:"omitempty,email"`
	Password             string `json:"password" validate:"required,min=8"`
	PasswordConfirmation string `json:"password_confirmation" validate:"required,eqfield=Password"`
	FullName             string `json:"full_name" validate:"required"`
	PhoneNumber          string `json:"phone_number"`

	RoleName      string `json:"role_name"`
	CaptchaToken  string `json:"captcha_token" validate:"required"`
	CaptchaAnswer string `json:"captcha_answer" validate:"required"`
}

type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required,min=6"`
}

type ResetPasswordRequest struct {
	Identifier      string `json:"identifier" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8"`
	HumanCheckToken string `json:"human_check_token" validate:"required"`
}

type Setup2FARequest struct {
	PendingToken string `json:"pending_token" validate:"required"`
}

type ConfirmSetup2FARequest struct {
	PendingToken string `json:"pending_token" validate:"required"`
	Secret       string `json:"secret" validate:"required"`
	OTPCode      string `json:"otp_code" validate:"required,len=6"`
}

type VerifyOTPRequest struct {
	PendingToken string `json:"pending_token" validate:"required"`
	OTPCode      string `json:"otp_code" validate:"required,len=6"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type UserSummary struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	RoleID   uint   `json:"role_id"`
	RoleName string `json:"role_name"`
}

type SessionInfo struct {
	ID             uint   `json:"id"`
	Browser        string `json:"browser"`
	BrowserVersion string `json:"browser_version"`
	OS             string `json:"os"`
	OSVersion      string `json:"os_version"`
	DeviceType     string `json:"device_type"`
	IPAddress      string `json:"ip_address"`
	Location       string `json:"location"`
	CreatedAt      string `json:"created_at"`
	IsCurrent      bool   `json:"is_current"`
}

type SessionListResponse struct {
	Sessions []SessionInfo `json:"sessions"`
}

type LoginResponse struct {
	RequireOTP   bool         `json:"require_otp"`
	PendingToken string       `json:"pending_token,omitempty"`
	TokenType    string       `json:"token_type,omitempty"`
	AccessToken  string       `json:"access_token,omitempty"`
	RefreshToken string       `json:"refresh_token,omitempty"`
	User         *UserSummary `json:"user,omitempty"`
	Session      *SessionInfo `json:"session,omitempty"`
}

type Setup2FAResponse struct {
	Secret    string `json:"secret"`
	QRCodePNG string `json:"qr_code_png"`
}

type MeResponse struct {
	ID           uint   `json:"id"`
	Username     string `json:"username"`
	Email        string `json:"email"`
	FullName     string `json:"full_name"`
	PhoneNumber  string `json:"phone_number"`
	RoleID       uint   `json:"role_id"`
	RoleName     string `json:"role_name"`
	Is2FAEnabled bool   `json:"is_2fa_enabled"`
	AvatarURL    string `json:"avatar_url"`
}
