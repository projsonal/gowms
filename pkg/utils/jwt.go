package utils

import (
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/projsonal/gostock/pkg/config"
)

type JWTClaims struct {
	UserID   uint   `json:"user_id"`
	RoleID   uint   `json:"role_id"`
	RoleName string `json:"role_name"`
	jwt.RegisteredClaims
}

type JWTService struct {
	cfg *config.JWTConfig
}

func NewJWTService(cfg *config.JWTConfig) *JWTService {
	return &JWTService{cfg: cfg}
}

func (s *JWTService) GenerateAccessToken(userID, roleID uint, roleName string) (string, error) {
	claims := JWTClaims{
		UserID:   userID,
		RoleID:   roleID,
		RoleName: roleName,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(s.cfg.AccessExpiryMinutes) * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.AccessSecret))
}

func (s *JWTService) GenerateRefreshToken(userID uint) (string, time.Time, error) {
	expiry := time.Now().AddDate(0, 0, s.cfg.RefreshExpiryDays)
	claims := jwt.RegisteredClaims{
		Subject:   strconv.FormatUint(uint64(userID), 10),
		ExpiresAt: jwt.NewNumericDate(expiry),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(s.cfg.RefreshSecret))
	return signed, expiry, err
}

func (s *JWTService) ParseAccessToken(tokenStr string) (*JWTClaims, error) {
	claims := &JWTClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(s.cfg.AccessSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("access token tidak valid")
	}
	return claims, nil
}

func (s *JWTService) ParseRefreshToken(tokenStr string) (*jwt.RegisteredClaims, error) {
	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(s.cfg.RefreshSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("refresh token tidak valid")
	}
	return claims, nil
}

func ParseUintSubject(subject string) uint {
	id, _ := strconv.ParseUint(subject, 10, 64)
	return uint(id)
}
