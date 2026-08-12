package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"github.com/projsonal/gowms/internal/middleware"
	"github.com/projsonal/gowms/internal/model"
	"github.com/projsonal/gowms/pkg/constant"
	"github.com/projsonal/gowms/pkg/utils"
)

const (
	maxFailedLoginAttempts = 10
	lockDuration           = 15 * time.Minute

	accountInactiveMessage = "Akun kamu terkunci, hubungi admin supaya bisa dibuka lagi akunnya."
	otpCodeReusedMessage   = "kode sudah tidak bisa digunakan, coba gunakan kode yang baru."
)

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func (h *Controller) buildSessionInfo(c *fiber.Ctx) (device utils.DeviceInfo, ip, location string) {
	device = utils.ParseUserAgent(c.Get("User-Agent"))
	ip = c.IP()

	location, err := h.geoipSvc.Lookup(ip)
	if err != nil {
		log.Printf("auth: lookup geoip untuk IP %s gagal: %v", ip, err)
		location = "-"
	}
	return device, ip, location
}

func (h *Controller) issueTokens(c *fiber.Ctx, u *model.User) (*LoginResponse, error) {
	r, err := h.roleRepo.FindByID(u.RoleID)
	if err != nil {
		return nil, err
	}

	refresh, expiry, err := h.jwtSvc.GenerateRefreshToken(u.ID)
	if err != nil {
		return nil, err
	}

	device, ip, location := h.buildSessionInfo(c)
	session := &model.RefreshToken{
		UserID:         u.ID,
		TokenHash:      hashToken(refresh),
		UserAgent:      c.Get("User-Agent"),
		Browser:        device.Browser,
		BrowserVersion: device.BrowserVersion,
		OS:             device.OS,
		OSVersion:      device.OSVersion,
		DeviceType:     string(device.DeviceType),
		IPAddress:      ip,
		Location:       location,
		ExpiresAt:      expiry,
	}
	if err := h.authRepo.SaveRefreshToken(session); err != nil {
		return nil, err
	}

	access, err := h.jwtSvc.GenerateAccessToken(u.ID, u.RoleID, session.ID, r.Name)
	if err != nil {
		return nil, err
	}
	if err := h.userRepo.UpdateLastLogin(u.ID); err != nil {
		log.Printf("auth: gagal update last_login_at untuk user %d: %v", u.ID, err)
	}

	return &LoginResponse{
		TokenType:    "Bearer",
		AccessToken:  access,
		RefreshToken: refresh,
		User: &UserSummary{
			ID: u.ID, Username: u.Username, Email: u.Email, RoleID: u.RoleID, RoleName: r.Name,
		},
		Session: &SessionInfo{
			ID:             session.ID,
			Browser:        device.Browser,
			BrowserVersion: device.BrowserVersion,
			OS:             device.OS,
			OSVersion:      device.OSVersion,
			DeviceType:     string(device.DeviceType),
			IPAddress:      ip,
			Location:       location,
			IsCurrent:      true,
		},
	}, nil
}

func (h *Controller) resolveRegisterRoleName(requestedRole string) string {
	if requestedRole == "" {
		return constant.RoleKaryawan
	}
	if h.appEnv == "production" {
		log.Printf("auth: percobaan self-register dengan role '%s' ditolak (APP_ENV=production), dipaksa ke '%s'", requestedRole, constant.RoleKaryawan)
		return constant.RoleKaryawan
	}
	return requestedRole
}

func (h *Controller) CheckUsernameAvailability(c *fiber.Ctx) error {
	username := strings.TrimSpace(c.Query("username"))
	if len(username) < 4 {
		return utils.OK(c, "username terlalu pendek", fiber.Map{"available": false})
	}
	_, err := h.userRepo.FindByUsername(username)
	available := err != nil // error (tidak ketemu) berarti TERSEDIA
	return utils.OK(c, "berhasil cek ketersediaan username", fiber.Map{"available": available})
}

func (h *Controller) Register(c *fiber.Ctx) error {
	var req RegisterRequest
	if !utils.ParseAndValidate(c, &req) {
		return nil
	}

	if req.CaptchaToken != "" || req.CaptchaAnswer != "" {
		if err := h.captchaSvc.Verify(req.CaptchaToken, req.CaptchaAnswer); err != nil {
			return utils.Fail(c, fiber.StatusBadRequest, "verifikasi captcha gagal: "+err.Error(), nil)
		}
	}

	if _, err := h.userRepo.FindByUsername(req.Username); err == nil {
		return utils.Fail(c, fiber.StatusConflict, "username sudah digunakan", nil)
	}
	if req.Email != "" {
		if _, err := h.userRepo.FindByEmail(req.Email); err == nil {
			return utils.Fail(c, fiber.StatusConflict, "email sudah terdaftar", nil)
		}
	}

	roleName := h.resolveRegisterRoleName(req.RoleName)
	targetRole, err := h.roleRepo.FindByName(roleName)
	if err != nil {
		log.Printf("auth: role '%s' belum ada di database, jalankan seeder role dulu: %v", roleName, err)
		return utils.Fail(c, fiber.StatusInternalServerError, "pendaftaran belum bisa diproses, hubungi administrator", nil)
	}

	hashed, err := utils.HashPassword(req.Password)
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal mengenkripsi password", nil)
	}

	newUser := &model.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashed,
		FullName:     req.FullName,
		PhoneNumber:  req.PhoneNumber,
		RoleID:       targetRole.ID,
		IsActive:     true,
		Is2FAEnabled: false,
	}
	if err := h.userRepo.Create(newUser); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal membuat akun", nil)
	}

	res, err := h.issueTokens(c, newUser)
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError,
			"akun berhasil dibuat, tapi gagal memulai sesi. Silakan login manual", nil)
	}
	return utils.Created(c, "akun berhasil dibuat", res)
}

func (h *Controller) ResetPassword(c *fiber.Ctx) error {
	var req ResetPasswordRequest
	if !utils.ParseAndValidate(c, &req) {
		return nil
	}

	if err := h.captchaSvc.Verify(req.CaptchaToken, req.CaptchaAnswer); err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "verifikasi captcha gagal: "+err.Error(), nil)
	}

	u, err := h.userRepo.FindByUsername(req.Identifier)
	if err != nil {
		u, err = h.userRepo.FindByEmail(req.Identifier)
	}
	if err != nil {
		return utils.Fail(c, fiber.StatusNotFound, "akun dengan username tersebut tidak ditemukan", nil)
	}
	if !u.IsActive {
		return utils.Fail(c, fiber.StatusForbidden, "akun anda tidak aktif, hubungi administrator", nil)
	}

	hashed, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal mengenkripsi password baru", nil)
	}
	u.PasswordHash = hashed
	if err := h.userRepo.Update(u); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal menyimpan password baru", nil)
	}

	if err := h.authRepo.RevokeAllUserTokens(u.ID); err != nil {
		log.Printf("auth: gagal mencabut sesi lama setelah reset password untuk user %d: %v", u.ID, err)
	}

	return utils.OK(c, "password berhasil diubah, silakan login dengan password baru", nil)
}

func (h *Controller) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if !utils.ParseAndValidate(c, &req) {
		return nil
	}

	u, err := h.userRepo.FindByUsername(req.Username)
	if h.loginAuthFailure(c, u, err, req.Password) {
		return nil
	}
	if h.loginLockedOrInactive(c, u) {
		return nil
	}
	if err := h.userRepo.ResetFailedLogin(u.ID); err != nil {
		log.Printf("auth: gagal reset failed_login_attempts untuk user %d: %v", u.ID, err)
	}

	if !u.Is2FAEnabled {
		res, err := h.issueTokens(c, u)
		if err != nil {
			return utils.Fail(c, fiber.StatusInternalServerError, "gagal memproses login", nil)
		}
		return utils.OK(c, "login berhasil", res)
	}

	pendingToken, _, err := h.jwtSvc.GenerateRefreshToken(u.ID)
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal memproses login", nil)
	}
	return utils.OK(c, "login berhasil, silakan verifikasi kode OTP", LoginResponse{
		RequireOTP: true, PendingToken: pendingToken,
	})
}

func (h *Controller) loginAuthFailure(c *fiber.Ctx, u *model.User, err error, password string) bool {
	if err != nil || !utils.ComparePassword(u.PasswordHash, password) {
		if u != nil {
			if err := h.userRepo.RegisterFailedLogin(u.ID, maxFailedLoginAttempts, lockDuration); err != nil {
				log.Printf("auth: gagal mencatat percobaan login gagal untuk user %d: %v", u.ID, err)
			}
		}
		_ = utils.Fail(c, fiber.StatusUnauthorized, "username atau password salah", nil)
		return true
	}
	return false
}

func (h *Controller) loginLockedOrInactive(c *fiber.Ctx, u *model.User) bool {
	if u.LockedUntil != nil && u.LockedUntil.After(time.Now()) {
		_ = utils.Fail(c, fiber.StatusTooManyRequests,
			"akun anda dikunci sementara karena terlalu banyak percobaan gagal, coba lagi setelah 15 menit", nil)
		return true
	}
	if !u.IsActive {
		_ = utils.Fail(c, fiber.StatusForbidden, accountInactiveMessage, nil)
		return true
	}
	return false
}

func (h *Controller) StartTwoFactorSetup(c *fiber.Ctx) error {
	userID, ok := c.Locals(constant.CtxUserID).(uint)
	if !ok {
		return utils.Fail(c, fiber.StatusUnauthorized, "sesi tidak valid", nil)
	}
	u, err := h.userRepo.FindByID(userID)
	if err != nil {
		return utils.Fail(c, fiber.StatusNotFound, constant.ErrUserNotFound, nil)
	}
	if u.Is2FAEnabled {
		return utils.Fail(c, fiber.StatusBadRequest, "2FA sudah aktif untuk akun ini", nil)
	}

	pendingToken, _, err := h.jwtSvc.GenerateRefreshToken(u.ID)
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal memulai setup 2FA", nil)
	}
	return utils.OK(c, "silakan lanjutkan setup 2FA", fiber.Map{"pending_token": pendingToken})
}

func (h *Controller) Setup2FA(c *fiber.Ctx) error {
	var req Setup2FARequest
	if !utils.ParseAndValidate(c, &req) {
		return nil
	}

	claims, err := h.jwtSvc.ParseRefreshToken(req.PendingToken)
	if err != nil {
		return utils.Fail(c, fiber.StatusUnauthorized, constant.ErrSessionExpired, nil)
	}

	u, err := h.userRepo.FindByID(utils.ParseUintSubject(claims.Subject))
	if err != nil {
		return utils.Fail(c, fiber.StatusNotFound, constant.ErrUserNotFound, nil)
	}

	totpLabel := u.Email
	if totpLabel == "" {
		totpLabel = u.Username
	}
	secret, qr, err := utils.GenerateTOTPSecret(h.totpIssuer, totpLabel)
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal membuat QR 2FA", nil)
	}
	return utils.OK(c, "silakan scan QR dengan Google Authenticator", Setup2FAResponse{
		Secret: secret, QRCodePNG: qr,
	})
}

func (h *Controller) ConfirmSetup2FA(c *fiber.Ctx) error {
	var req ConfirmSetup2FARequest
	if !utils.ParseAndValidate(c, &req) {
		return nil
	}

	claims, err := h.jwtSvc.ParseRefreshToken(req.PendingToken)
	if err != nil {
		return utils.Fail(c, fiber.StatusUnauthorized, constant.ErrSessionExpired, nil)
	}
	userID := utils.ParseUintSubject(claims.Subject)

	if !utils.VerifyTOTP(req.OTPCode, req.Secret) {
		return utils.Fail(c, fiber.StatusBadRequest, "Kode OTP salah atau sudah kedaluwarsa", nil)
	}
	if !h.otpReplayGuard.checkAndMark(userID, req.OTPCode) {
		return utils.Fail(c, fiber.StatusBadRequest, otpCodeReusedMessage, nil)
	}
	if err := h.userRepo.UpdateTOTPSecret(userID, req.Secret, true); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal mengaktifkan 2FA", nil)
	}

	u, err := h.userRepo.FindByID(userID)
	if err != nil {
		return utils.Fail(c, fiber.StatusNotFound, constant.ErrUserNotFound, nil)
	}
	res, err := h.issueTokens(c, u)
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal menerbitkan sesi", nil)
	}
	return utils.OK(c, "Two Factor Authentication berhasil diaktifkan", res)
}

func (h *Controller) verifyOTPCode(req VerifyOTPRequest, totpSecret string) bool {
	return utils.VerifyTOTP(req.OTPCode, totpSecret)
}

func (h *Controller) VerifyOTP(c *fiber.Ctx) error {
	var req VerifyOTPRequest
	if !utils.ParseAndValidate(c, &req) {
		return nil
	}

	claims, err := h.jwtSvc.ParseRefreshToken(req.PendingToken)
	if err != nil {
		return utils.Fail(c, fiber.StatusUnauthorized, constant.ErrSessionExpired, nil)
	}

	userID := utils.ParseUintSubject(claims.Subject)
	u, err := h.userRepo.FindByID(userID)
	if err != nil {
		return utils.Fail(c, fiber.StatusNotFound, constant.ErrUserNotFound, nil)
	}
	if !h.verifyOTPCode(req, u.TOTPSecret) {
		return utils.Fail(c, fiber.StatusBadRequest, "Kode OTP salah atau sudah kedaluwarsa", nil)
	}
	if !h.otpReplayGuard.checkAndMark(u.ID, req.OTPCode) {
		return utils.Fail(c, fiber.StatusBadRequest, "Kode OTP ini sudah pernah dipakai, silakan gunakan kode baru", nil)
	}

	res, err := h.issueTokens(c, u)
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal menerbitkan sesi", nil)
	}
	return utils.OK(c, "Verifikasi OTP Berhasil", res)
}

func (h *Controller) RefreshToken(c *fiber.Ctx) error {
	var req RefreshTokenRequest
	if !utils.ParseAndValidate(c, &req) {
		return nil
	}

	claims, err := h.jwtSvc.ParseRefreshToken(req.RefreshToken)
	if err != nil {
		return utils.Fail(c, fiber.StatusUnauthorized, "refresh token tidak valid atau kedaluwarsa", nil)
	}

	userID := utils.ParseUintSubject(claims.Subject)
	if _, err := h.authRepo.FindActiveRefreshToken(userID, hashToken(req.RefreshToken)); err != nil {
		return utils.Fail(c, fiber.StatusUnauthorized, "sesi tidak ditemukan atau sudah di-revoke", nil)
	}

	u, err := h.userRepo.FindByID(userID)
	if err != nil {
		return utils.Fail(c, fiber.StatusNotFound, constant.ErrUserNotFound, nil)
	}

	res, err := h.issueTokens(c, u)
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal memperbarui token", nil)
	}
	return utils.OK(c, "token berhasil diperbarui", res)
}

func (h *Controller) Logout(c *fiber.Ctx) error {
	userID, _ := c.Locals(constant.CtxUserID).(uint)
	if err := h.authRepo.RevokeAllUserTokens(userID); err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal logout", nil)
	}
	return utils.OK(c, "logout berhasil", nil)
}

func toSessionInfo(s model.RefreshToken, currentSessionID uint) SessionInfo {
	return SessionInfo{
		ID:             s.ID,
		Browser:        s.Browser,
		BrowserVersion: s.BrowserVersion,
		OS:             s.OS,
		OSVersion:      s.OSVersion,
		DeviceType:     s.DeviceType,
		IPAddress:      s.IPAddress,
		Location:       s.Location,
		CreatedAt:      s.CreatedAt.Format(time.RFC3339),
		IsCurrent:      currentSessionID != 0 && s.ID == currentSessionID,
	}
}

func (h *Controller) ListSessions(c *fiber.Ctx) error {
	userID, ok := c.Locals(constant.CtxUserID).(uint)
	if !ok {
		return utils.Fail(c, fiber.StatusUnauthorized, constant.ErrInvalidToken, nil)
	}
	currentSessionID, _ := c.Locals(constant.CtxSessionID).(uint)

	sessions, err := h.authRepo.ListActiveSessions(userID)
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal memuat daftar sesi", nil)
	}

	result := make([]SessionInfo, 0, len(sessions))
	for _, s := range sessions {
		result = append(result, toSessionInfo(s, currentSessionID))
	}
	return utils.OK(c, "berhasil memuat daftar sesi aktif", SessionListResponse{Sessions: result})
}

func (h *Controller) RevokeSession(c *fiber.Ctx) error {
	userID, ok := c.Locals(constant.CtxUserID).(uint)
	if !ok {
		return utils.Fail(c, fiber.StatusUnauthorized, constant.ErrInvalidToken, nil)
	}

	var params model.IDParam
	if err := c.ParamsParser(&params); err != nil {
		return utils.Fail(c, fiber.StatusBadRequest, "id sesi tidak valid", nil)
	}

	if err := h.authRepo.RevokeSession(userID, params.ID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.Fail(c, fiber.StatusNotFound, "sesi tidak ditemukan", nil)
		}
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal mencabut sesi", nil)
	}

	currentSessionID, _ := c.Locals(constant.CtxSessionID).(uint)
	revokedCurrent := currentSessionID != 0 && currentSessionID == params.ID
	return utils.OK(c, "sesi berhasil dicabut", fiber.Map{"revoked_current": revokedCurrent})
}

func (h *Controller) Me(c *fiber.Ctx) error {
	userID, ok := c.Locals(constant.CtxUserID).(uint)
	if !ok {
		return utils.Fail(c, fiber.StatusUnauthorized, constant.ErrInvalidToken, nil)
	}

	u, err := h.userRepo.FindByID(userID)
	if err != nil {
		return utils.Fail(c, fiber.StatusNotFound, constant.ErrUserNotFound, nil)
	}
	r, err := h.roleRepo.FindByID(u.RoleID)
	if err != nil {
		return utils.Fail(c, fiber.StatusInternalServerError, "gagal memuat data role", nil)
	}

	return utils.OK(c, "berhasil membaca sesi aktif", MeResponse{
		ID:           u.ID,
		Username:     u.Username,
		Email:        u.Email,
		FullName:     u.FullName,
		PhoneNumber:  u.PhoneNumber,
		RoleID:       u.RoleID,
		RoleName:     r.Name,
		Is2FAEnabled: u.Is2FAEnabled,
	})
}

func (h *Controller) RegisterRoutes(router fiber.Router) {
	g := router.Group("/auth")

	g.Get("/username-available", middleware.UsernameCheckRateLimiter(), h.CheckUsernameAvailability)
	g.Post("/register", middleware.RegisterRateLimiter(), h.Register)
	g.Post("/login", h.Login)
	g.Post("/2fa/setup", h.Setup2FA)
	g.Post("/2fa/confirm", h.ConfirmSetup2FA)
	g.Post("/verify-otp", h.VerifyOTP)
	g.Post("/refresh", h.RefreshToken)

	g.Post("/password/reset", middleware.PasswordResetRateLimiter(), h.ResetPassword)

	protected := g.Group("/", middleware.JWTAuth(h.jwtSvc))
	protected.Post("/logout", h.Logout)
	protected.Get("/me", h.Me)
	protected.Post("/2fa/start", h.StartTwoFactorSetup)
	protected.Get("/sessions", h.ListSessions)
	protected.Delete("/sessions/:id", h.RevokeSession)
}
