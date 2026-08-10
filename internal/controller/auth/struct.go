package auth

import (
	"time"

	authRepo "github.com/projsonal/gowms/internal/repositories/auth"
	"github.com/projsonal/gowms/internal/repositories/role"
	"github.com/projsonal/gowms/internal/repositories/users"
	"github.com/projsonal/gowms/pkg/captcha"
	"github.com/projsonal/gowms/pkg/config"
	"github.com/projsonal/gowms/pkg/geoip"
	"github.com/projsonal/gowms/pkg/utils"
)

const (
	// (Konstanta metode OTP WhatsApp/SMS sudah dihapus seluruhnya —
	// login, registrasi, ganti password, dan reset password sekarang
	// semua memakai captcha, bukan lagi OTP via WhatsApp/SMS.)

	// totpReplayWindow: seberapa lama sebuah kode TOTP yang SUDAH DIPAKAI
	// ditolak kalau dicoba lagi (lihat totp_replay_guard.go). 90 detik
	// dipilih supaya menutupi jendela toleransi clock-skew ±1 langkah
	// (30 detik) yang diizinkan pkg/utils.VerifyTOTP, jadi kode yang
	// sama tidak bisa dipakai dua kali walau masih dalam toleransi itu.
	totpReplayWindow = 90 * time.Second
)

type Controller struct {
	authRepo   authRepo.Repository
	userRepo   users.Repository
	roleRepo   role.Repository
	jwtSvc     *utils.JWTService
	captchaSvc *captcha.Service
	totpIssuer string
	appEnv     string

	geoipSvc geoip.Resolver

	// otpReplayGuard mencegah satu kode TOTP dipakai berkali-kali dalam
	// jendela validitasnya — tanpa ini, kode 2FA yang bocor/tersadap bisa
	// dipakai berulang selama masih dalam 30 detik masa berlakunya.
	otpReplayGuard *totpReplayGuard
}

type Params struct {
	AuthRepo   authRepo.Repository
	UserRepo   users.Repository
	RoleRepo   role.Repository
	JWTSvc     *utils.JWTService
	CaptchaSvc *captcha.Service
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
		totpIssuer: p.Cfg.TOTP.Issuer,
		appEnv:     p.Cfg.App.Env,
		geoipSvc:   p.GeoipSvc,

		otpReplayGuard: newTOTPReplayGuard(totpReplayWindow),
	}
}

type RegisterRequest struct {
	Username             string `json:"username" validate:"required,min=4,max=50"`
	// Email OPSIONAL — form registrasi (lihat RegisterStep di frontend)
	// sudah tidak mengumpulkan email sama sekali; identitas login pakai
	// username. Field dipertahankan (bukan dihapus) supaya kontrak API
	// tidak breaking kalau suatu saat email diaktifkan lagi, dan supaya
	// endpoint lama yang masih mengirim email tetap tervalidasi benar.
	Email                string `json:"email" validate:"omitempty,email"`
	Password             string `json:"password" validate:"required,min=8"`
	PasswordConfirmation string `json:"password_confirmation" validate:"required,eqfield=Password"`
	FullName             string `json:"full_name" validate:"required"`
	// Nomor HP WAJIB diisi (bukan lagi opsional) karena sekarang dipakai
	// untuk verifikasi OTP saat registrasi (lihat RequestRegisterOTP /
	// VerifyRegisterOTP di bawah) — tanpa nomor valid, langkah verifikasi
	// tidak mungkin dijalankan.
	// Nomor HP OPSIONAL. Verifikasi OTP saat registrasi (WA/SMS) sudah
	// dilepas dari alur wajib karena gateway pengirimnya sering belum
	// terkonfigurasi di lingkungan dev — 2FA (Google Authenticator) tetap
	// jadi satu-satunya lapisan verifikasi wajib untuk aktivasi akun.
	// Endpoint /auth/register/otp/request & /otp/verify TETAP ada untuk
	// dipakai lagi nanti kalau gateway WA/SMS sudah siap.
	PhoneNumber   string `json:"phone_number" validate:"omitempty,e164"`
	RoleName      string `json:"role_name" validate:"omitempty,oneof=super_admin admin karyawan analis_data"`
	// Captcha OPSIONAL: aplikasi ini khusus internal perusahaan (bukan
	// pendaftaran publik terbuka), jadi UI captcha sengaja tidak
	// ditampilkan lagi di frontend. Kalau frontend TETAP mengirim
	// captcha_token/captcha_answer (mis. suatu saat captcha diaktifkan
	// lagi untuk skenario lain), field-nya tetap diverifikasi di bawah —
	// jadi endpoint ini tetap aman dipakai kalau captcha diaktifkan lagi
	// nanti tanpa perlu ubah kontrak API.
	CaptchaToken  string `json:"captcha_token" validate:"omitempty"`
	CaptchaAnswer string `json:"captcha_answer" validate:"omitempty"`
}

type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required,min=6"`
}

type LoginResponse struct {
	RequirePhoneVerification bool         `json:"require_phone_verification"`
	RequireSetup2FA          bool         `json:"require_setup_2fa"`
	RequireOTP               bool         `json:"require_otp"`
	PendingToken             string       `json:"pending_token,omitempty"`
	TokenType                string       `json:"token_type,omitempty"`
	AccessToken              string       `json:"access_token,omitempty"`
	RefreshToken             string       `json:"refresh_token,omitempty"`
	User                     *UserSummary `json:"user,omitempty"`
	Session                  *SessionInfo `json:"session,omitempty"`
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
	// IsCurrent: true kalau ini sesi yang sedang dipakai request ini
	// sendiri (lihat JWTClaims.SessionID) — dipakai frontend menandai
	// "Perangkat ini" & memicu logout otomatis kalau di-Cabut sendiri.
	IsCurrent bool `json:"is_current"`
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
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// ---------------------------------------------------------------------

// ResetPasswordRequest — reset password langsung dalam SATU langkah (TIDAK
// lagi lewat OTP WhatsApp/SMS), diverifikasi captcha gambar self-hosted.
//
// Catatan keamanan yang WAJIB diketahui: menghapus verifikasi kepemilikan
// akun (OTP ke nomor terdaftar) berarti SIAPA PUN yang tahu username/email
// sebuah akun kini bisa mengganti passwordnya HANYA dengan menyelesaikan
// captcha — captcha cuma membuktikan "bukan bot", BUKAN "pemilik akun
// ini". Ini keputusan produk yang diminta eksplisit (menyederhanakan alur
// lupa password), bukan kelalaian — tapi risikonya nyata: akun bisa
// diambil alih orang lain yang tahu/menebak username. Mitigasi minimal
// yang tetap dipertahankan: PasswordResetRateLimiter di RegisterRoutes.
type ResetPasswordRequest struct {
	// Identifier boleh username ATAU email.
	Identifier              string `json:"identifier" validate:"required"`
	NewPassword             string `json:"new_password" validate:"required,min=8"`
	NewPasswordConfirmation string `json:"new_password_confirmation" validate:"required,eqfield=NewPassword"`
	CaptchaToken            string `json:"captcha_token" validate:"required"`
	CaptchaAnswer           string `json:"captcha_answer" validate:"required"`
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
