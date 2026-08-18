package humancheck

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"time"
)

var (
	ErrInvalid     = errors.New("token verifikasi tidak valid")
	ErrExpired     = errors.New("token verifikasi sudah kadaluarsa, silakan verifikasi ulang")
	ErrTooFast     = errors.New("Verifikasi terlalu cepat, tunggu waktu sesaat kemudian coba lagi")
	ErrAlreadyUsed = errors.New("token verifikasi sudah digunakan, silakan verifikasi ulang")
)

// Service memverifikasi human-check token: HMAC-signed timestamp + nonce
// yang harus di-"issue" lebih dulu, tidak boleh dipakai lebih cepat dari
// minDelay (menggagalkan bot yang submit instan), tidak boleh kedaluwarsa
// dari ttl, dan tidak boleh dipakai dua kali (replay).
type Service struct {
	secret   []byte
	ttl      time.Duration
	minDelay time.Duration
	used     *usedStore
}

func NewService(secret string, ttl, minDelay time.Duration) *Service {
	return &Service{
		secret:   []byte(secret),
		ttl:      ttl,
		minDelay: minDelay,
		used:     newUsedStore(),
	}
}

func (s *Service) Issue() (string, error) {
	issuedAt := time.Now().Unix()

	payload := make([]byte, 8, 16)
	binary.BigEndian.PutUint64(payload[0:8], uint64(issuedAt))

	nonce := make([]byte, 8)
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	payload = append(payload, nonce...)

	mac := hmac.New(sha256.New, s.secret)
	mac.Write(payload)
	sig := mac.Sum(nil)

	return base64.RawURLEncoding.EncodeToString(append(payload, sig...)), nil
}

func (s *Service) Verify(token string) error {
	issuedAt, err := s.parseToken(token)
	if err != nil {
		return err
	}
	if s.used.isUsed(token) {
		return ErrAlreadyUsed
	}

	elapsed := time.Since(issuedAt)
	if elapsed > s.ttl {
		return ErrExpired
	}
	s.used.markUsed(token, s.ttl)

	if elapsed < s.minDelay {
		return ErrTooFast
	}
	return nil
}

func (s *Service) parseToken(token string) (issuedAt time.Time, err error) {
	raw, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil || len(raw) != 16+sha256.Size {
		return time.Time{}, ErrInvalid
	}

	payload := raw[:16]
	sig := raw[16:]

	mac := hmac.New(sha256.New, s.secret)
	mac.Write(payload)
	if !hmac.Equal(sig, mac.Sum(nil)) {
		return time.Time{}, ErrInvalid
	}

	sec := int64(binary.BigEndian.Uint64(payload[0:8]))
	return time.Unix(sec, 0), nil
}
