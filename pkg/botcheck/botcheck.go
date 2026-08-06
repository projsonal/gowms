package botcheck

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"time"
)

var ErrInvalidOrExpired = errors.New("sesi verifikasi kedaluwarsa atau tidak valid")

type Service struct {
	secret []byte
	window time.Duration
}

func NewService(secret string, window time.Duration) *Service {
	return &Service{secret: []byte(secret), window: window}
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

func (s *Service) Verify(token string) bool {
	issuedAt, ok := s.parseToken(token)
	if !ok {
		return false
	}
	return time.Since(issuedAt) <= s.window
}

func (s *Service) parseToken(token string) (issuedAt time.Time, ok bool) {
	raw, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil || len(raw) != 16+sha256.Size {
		return time.Time{}, false
	}

	payload := raw[:16]
	sig := raw[16:]

	mac := hmac.New(sha256.New, s.secret)
	mac.Write(payload)
	if !hmac.Equal(sig, mac.Sum(nil)) {
		return time.Time{}, false
	}

	sec := int64(binary.BigEndian.Uint64(payload[0:8]))
	return time.Unix(sec, 0), true
}
