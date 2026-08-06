package auth

import (
	"time"

	authRepo "github.com/projsonal/gowms/internal/repositories/auth"
	"github.com/projsonal/gowms/internal/repositories/role"
	"github.com/projsonal/gowms/internal/repositories/users"
	"github.com/projsonal/gowms/pkg/captcha"
	"github.com/projsonal/gowms/pkg/config"
	"github.com/projsonal/gowms/pkg/geoip"
	"github.com/projsonal/gowms/pkg/otp"
	"github.com/projsonal/gowms/pkg/utils"
	"github.com/projsonal/gowms/pkg/wa"
)

const (
	MethodTOTP     = "totp"
	MethodWhatsApp = "whatsapp"
)

type Controller struct {
	authRepo   authRepo.Repository
	userRepo   users.Repository
	roleRepo   role.Repository
	jwtSvc     *utils.JWTService
	captchaSvc *captcha.Service
	waOTPSvc   *otp.Service
	waSender   wa.Sender
	waOTPTTL   time.Duration
	totpIssuer string
	appEnv     string

	geoipSvc geoip.Resolver
}

type Params struct {
	AuthRepo   authRepo.Repository
	UserRepo   users.Repository
	RoleRepo   role.Repository
	JWTSvc     *utils.JWTService
	CaptchaSvc *captcha.Service
	WaOTPSvc   *otp.Service
	WaSender   wa.Sender
	Cfg        *config.Config
	GeoipSvc   geoip.Resolver
}

func New(p Params) *Controller {
	return &Controller{
		authRepo:   p.AuthRepo,
		userRepo:   p.UserRepo,
		roleRepo:   p.RoleRepo,
		jwtSvc:     p.JWTSvc,
		captchaSvc: p.CaptchaSvc,
		waOTPSvc:   p.WaOTPSvc,
		waSender:   p.WaSender,
		waOTPTTL:   time.Duration(p.Cfg.WAOTP.TTLMinutes) * time.Minute,
		totpIssuer: p.Cfg.TOTP.Issuer,
		appEnv:     p.Cfg.App.Env,
		geoipSvc:   p.GeoipSvc,
	}
}

type RegisterRequest struct {
	Username             string `json:"username" validate:"required,min=4,max=50"`
	Email                string `json:"email" validate:"required,email"`
	Password             string `json:"password" validate:"required,min=8"`
	PasswordConfirmation string `json:"password_confirmation" validate:"required,eqfield=Password"`
	FullName             string `json:"full_name" validate:"required"`
	PhoneNumber          string `json:"phone_number" validate:"omitempty,e164"`
	RoleName             string `json:"role_name" validate:"omitempty,oneof=super_admin admin karyawan analis_data"`
	CaptchaToken         string `json:"captcha_token" validate:"required"`
	CaptchaAnswer        string `json:"captcha_answer" validate:"required"`
}

type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required,min=6"`
}

type LoginResponse struct {
	RequireSetup2FA bool         `json:"require_setup_2fa"`
	RequireOTP      bool         `json:"require_otp"`
	PendingToken    string       `json:"pending_token,omitempty"`
	TokenType       string       `json:"token_type,omitempty"`
	AccessToken     string       `json:"access_token,omitempty"`
	RefreshToken    string       `json:"refresh_token,omitempty"`
	User            *UserSummary `json:"user,omitempty"`
	Session         *SessionInfo `json:"session,omitempty"`
}

type SessionInfo struct {
	ID             uint   `json:"id,omitempty"`
	Browser        string `json:"browser"`
	BrowserVersion string `json:"browser_version"`
	OS             string `json:"os"`
	OSVersion      string `json:"os_version"`
	DeviceType     string `json:"device_type"`
	IPAddress      string `json:"ip_address"`
	Location       string `json:"location"`
	CreatedAt      string `json:"created_at,omitempty"`
}

type SessionListResponse struct {
	Sessions []SessionInfo `json:"sessions"`
}

type UserSummary struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	RoleID   uint   `json:"role_id"`
	RoleName string `json:"role_name"`
}

type Setup2FAResponse struct {
	Secret    string `json:"secret"`
	QRCodePNG string `json:"qr_code_png_base64"`
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
	Method       string `json:"method" validate:"omitempty,oneof=totp whatsapp"`
	OTPToken     string `json:"otp_token" validate:"required_if=Method whatsapp"`
}

type RequestOTPRequest struct {
	PendingToken string `json:"pending_token" validate:"required"`
	Method       string `json:"method" validate:"required,oneof=whatsapp"`
}

type RequestOTPResponse struct {
	OTPToken  string `json:"otp_token"`
	ExpiresIn int    `json:"expires_in_seconds"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
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
}
