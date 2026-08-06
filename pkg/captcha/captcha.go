package captcha

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	mrand "math/rand"
	"strconv"
	"strings"
	"time"
)

var (
	ErrExpired     = errors.New("captcha sudah kedaluwarsa, silakan minta ulang")
	ErrInvalid     = errors.New("token captcha tidak valid")
	ErrWrongAnswer = errors.New("jawaban captcha salah")
	ErrAlreadyUsed = errors.New("captcha sudah pernah dipakai, silakan minta ulang")
)

type Service struct {
	secret []byte
	ttl    time.Duration
	used   *usedStore
}

func NewService(secret string, ttl time.Duration) *Service {
	return &Service{
		secret: []byte(secret),
		ttl:    ttl,
		used:   newUsedStore(),
	}
}

type Challenge struct {
	Token       string `json:"captcha_token"`
	ImageBase64 string `json:"captcha_image_base64"` // "data:image/png;base64,...."
}

func (s *Service) Generate() (*Challenge, error) {
	a := mrand.Intn(9) + 1
	b := mrand.Intn(9) + 1
	ops := []rune{'+', '-', 'x'}
	op := ops[mrand.Intn(len(ops))]

	var answer int
	switch op {
	case '+':
		answer = a + b
	case '-':
		if a < b {
			a, b = b, a
		}
		answer = a - b
	case 'x':
		answer = a * b
	}

	question := fmt.Sprintf("%d %c %d = ?", a, op, b)

	img := renderCaptchaImage(question)
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}

	token, err := s.signToken(answer)
	if err != nil {
		return nil, err
	}

	return &Challenge{
		Token:       token,
		ImageBase64: "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes()),
	}, nil
}

func (s *Service) Verify(token, userAnswer string) error {
	answer, expiresAt, err := s.parseToken(token)
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

	given, err := strconv.Atoi(strings.TrimSpace(userAnswer))
	if err != nil || given != answer {
		return ErrWrongAnswer
	}
	return nil
}

// ---- token signing (HMAC-SHA256, stateless & tamper-proof) ----

func (s *Service) signToken(answer int) (string, error) {
	expiresAt := time.Now().Add(s.ttl).Unix()

	payload := make([]byte, 12)
	binary.BigEndian.PutUint32(payload[0:4], uint32(answer))
	binary.BigEndian.PutUint64(payload[4:12], uint64(expiresAt))

	nonce := make([]byte, 8)
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	payload = append(payload, nonce...)

	mac := hmac.New(sha256.New, s.secret)
	mac.Write(payload)
	sig := mac.Sum(nil)

	full := append(payload, sig...)
	return base64.RawURLEncoding.EncodeToString(full), nil
}

func (s *Service) parseToken(token string) (answer int, expiresAt time.Time, err error) {
	raw, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil || len(raw) != 12+8+sha256.Size {
		return 0, time.Time{}, ErrInvalid
	}

	payload := raw[:20]
	sig := raw[20:]

	mac := hmac.New(sha256.New, s.secret)
	mac.Write(payload)
	expectedSig := mac.Sum(nil)
	if !hmac.Equal(sig, expectedSig) {
		return 0, time.Time{}, ErrInvalid
	}

	a := int32(binary.BigEndian.Uint32(payload[0:4]))
	exp := int64(binary.BigEndian.Uint64(payload[4:12]))
	return int(a), time.Unix(exp, 0), nil
}

// ---- render gambar PNG ----

func renderCaptchaImage(text string) image.Image {
	const (
		charW, charH = 5, 7
		scale        = 4
		padding      = 12
		gap          = 6
	)
	width := padding*2 + len(text)*(charW*scale+gap)
	height := padding*2 + charH*scale

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	bg := color.RGBA{245, 245, 248, 255}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, bg)
		}
	}

	for i := 0; i < 5; i++ {
		drawNoiseLine(img, width, height)
	}

	for i := 0; i < 60; i++ {
		x, y := mrand.Intn(width), mrand.Intn(height)
		img.Set(x, y, randomGray())
	}

	x := padding
	for _, ch := range text {
		glyph, ok := font5x7[ch]
		if !ok {
			glyph = font5x7[' ']
		}
		yOffset := mrand.Intn(5) - 2 // distorsi vertikal ringan per karakter
		drawGlyph(img, glyph, x, padding+yOffset, scale, randomDarkColor())
		x += charW*scale + gap
	}

	return img
}

func drawGlyph(img *image.RGBA, glyph [7]byte, ox, oy, scale int, col color.Color) {
	for row := 0; row < 7; row++ {
		bits := glyph[row]
		for col2 := 0; col2 < 5; col2++ {
			if bits&(1<<uint(4-col2)) != 0 {
				for sy := 0; sy < scale; sy++ {
					for sx := 0; sx < scale; sx++ {
						img.Set(ox+col2*scale+sx, oy+row*scale+sy, col)
					}
				}
			}
		}
	}
}

func drawNoiseLine(img *image.RGBA, width, height int) {
	x1, y1 := mrand.Intn(width), mrand.Intn(height)
	x2, y2 := mrand.Intn(width), mrand.Intn(height)
	col := randomGray()

	steps := 100
	for i := 0; i <= steps; i++ {
		t := float64(i) / float64(steps)
		x := int(float64(x1) + t*float64(x2-x1))
		y := int(float64(y1) + t*float64(y2-y1))
		if x >= 0 && x < width && y >= 0 && y < height {
			img.Set(x, y, col)
		}
	}
}

func randomGray() color.RGBA {
	v := uint8(160 + mrand.Intn(60))
	return color.RGBA{v, v, v, 255}
}

func randomDarkColor() color.RGBA {
	palette := []color.RGBA{
		{30, 41, 59, 255},
		{55, 48, 163, 255},
		{15, 118, 110, 255},
		{136, 19, 55, 255},
	}
	return palette[mrand.Intn(len(palette))]
}
