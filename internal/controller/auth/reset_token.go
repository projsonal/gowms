package auth

import (
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	resetStageRequested = "requested" // OTP baru dikirim, belum diverifikasi
	resetStageVerified  = "verified"  // OTP sudah benar, boleh set password baru
)

var (
	errResetTokenInvalid = errors.New("sesi reset password tidak valid, silakan ulangi dari awal")
	errResetStageInvalid = errors.New("langkah reset password tidak berurutan, silakan ulangi dari awal")
)

type resetClaims struct {
	UserID   uint   `json:"user_id"`
	Stage    string `json:"stage"`
	OTPToken string `json:"otp_token,omitempty"`
	Method   string `json:"method,omitempty"`
	jwt.RegisteredClaims
}

// resetTokenService menandatangani & memverifikasi token sesi lupa
// password. Sengaja terpisah dari utils.JWTService (access/refresh token
// login) karena secret & masa berlakunya beda konteks sepenuhnya.
type resetTokenService struct {
	secret []byte
	ttl    time.Duration
}

func newResetTokenService(secret string, ttl time.Duration) *resetTokenService {
	return &resetTokenService{secret: []byte(secret), ttl: ttl}
}

func (s *resetTokenService) generate(userID uint, stage, otpToken, method string) (string, error) {
	claims := resetClaims{
		UserID:   userID,
		Stage:    stage,
		OTPToken: otpToken,
		Method:   method,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatUint(uint64(userID), 10),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

func (s *resetTokenService) parse(tokenStr string) (*resetClaims, error) {
	claims := &resetClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		return s.secret, nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil || !token.Valid {
		return nil, errResetTokenInvalid
	}
	return claims, nil
}
