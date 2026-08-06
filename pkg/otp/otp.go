package otp

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	mrand "math/rand"
	"time"
)

var (
	ErrInvalid     = errors.New("token OTP tidak valid")
	ErrExpired     = errors.New("kode OTP sudah kedaluwarsa, silakan minta ulang")
	ErrWrongCode   = errors.New("kode OTP salah")
	ErrAlreadyUsed = errors.New("kode OTP sudah pernah dipakai, silakan minta ulang")
)

type Service struct {
	secret []byte
	ttl    time.Duration
	used   *usedStore
}

func NewService(secret string, ttl time.Duration) *Service {
	return &Service{secret: []byte(secret), ttl: ttl, used: newUsedStore()}
}

func (s *Service) Generate() (code string, token string, err error) {
	code = fmt.Sprintf("%06d", mrand.Intn(1_000_000))

	expiresAt := time.Now().Add(s.ttl).Unix()
	codeHash := hashCode(code)

	payload := make([]byte, 8, 8+len(codeHash))
	binary.BigEndian.PutUint64(payload[0:8], uint64(expiresAt))
	payload = append(payload, codeHash...)

	mac := hmac.New(sha256.New, s.secret)
	mac.Write(payload)
	sig := mac.Sum(nil)

	token = base64.RawURLEncoding.EncodeToString(append(payload, sig...))
	return code, token, nil
}

func (s *Service) Verify(token, userCode string) error {
	expiresAt, codeHash, err := s.parseToken(token)
	if err != nil {
		return err
	}

	if s.used.isUsed(token) {
		return ErrAlreadyUsed
	}
	s.used.markUsed(token, s.ttl)

	if time.Now().After(expiresAt) {
		return ErrExpired
	}
	if subtle.ConstantTimeCompare(hashCode(userCode), codeHash) != 1 {
		return ErrWrongCode
	}
	return nil
}

const headerLen = 8

func (s *Service) parseToken(token string) (expiresAt time.Time, codeHash []byte, err error) {
	raw, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil || len(raw) != headerLen+sha256.Size+sha256.Size {
		return time.Time{}, nil, ErrInvalid
	}

	payload := raw[:headerLen+sha256.Size]
	sig := raw[headerLen+sha256.Size:]

	mac := hmac.New(sha256.New, s.secret)
	mac.Write(payload)
	if !hmac.Equal(sig, mac.Sum(nil)) {
		return time.Time{}, nil, ErrInvalid
	}

	exp := int64(binary.BigEndian.Uint64(payload[0:8]))
	return time.Unix(exp, 0), payload[8:], nil
}

func hashCode(code string) []byte {
	sum := sha256.Sum256([]byte(code))
	return sum[:]
}
