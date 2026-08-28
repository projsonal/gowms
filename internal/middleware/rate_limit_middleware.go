package middleware

import (
	"os"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/projsonal/gowms/pkg/utils"
)

func devMultiplier() int {
	if os.Getenv("APP_ENV") == "production" {
		return 1
	}
	return 8
}

type ipRateLimiter struct {
	mu     sync.Mutex
	hits   map[string][]time.Time
	max    int
	window time.Duration
}

func newIPRateLimiter(max int, window time.Duration) *ipRateLimiter {
	rl := &ipRateLimiter{
		hits:   make(map[string][]time.Time),
		max:    max,
		window: window,
	}
	go rl.cleanupLoop()
	return rl
}

func (rl *ipRateLimiter) allow(key string) bool {
	now := time.Now()
	cutoff := now.Add(-rl.window)

	rl.mu.Lock()
	defer rl.mu.Unlock()

	hits := rl.hits[key]
	fresh := hits[:0]
	for _, t := range hits {
		if t.After(cutoff) {
			fresh = append(fresh, t)
		}
	}

	if len(fresh) >= rl.max {
		rl.hits[key] = fresh
		return false
	}

	rl.hits[key] = append(fresh, now)
	return true
}

func (rl *ipRateLimiter) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		cutoff := time.Now().Add(-rl.window)
		rl.mu.Lock()
		for key, hits := range rl.hits {
			fresh := hits[:0]
			for _, t := range hits {
				if t.After(cutoff) {
					fresh = append(fresh, t)
				}
			}
			if len(fresh) == 0 {
				delete(rl.hits, key)
			} else {
				rl.hits[key] = fresh
			}
		}
		rl.mu.Unlock()
	}
}

const (
	loginRateLimitMax    = 10
	loginRateLimitWindow = time.Minute
)

var loginLimiter = newIPRateLimiter(loginRateLimitMax, loginRateLimitWindow)

func LoginRateLimiter() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if !loginLimiter.allow(c.IP()) {
			return utils.Fail(c, fiber.StatusTooManyRequests,
				"terlalu banyak percobaan login dari alamat ini, coba lagi sebentar lagi", nil)
		}
		return c.Next()
	}
}

const (
	registerRateLimitMax    = 5
	registerRateLimitWindow = 10 * time.Minute
)

var registerLimiter = newIPRateLimiter(registerRateLimitMax*devMultiplier(), registerRateLimitWindow)

func RegisterRateLimiter() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if !registerLimiter.allow(c.IP()) {
			return utils.Fail(c, fiber.StatusTooManyRequests,
				"terlalu banyak percobaan pendaftaran dari alamat ini, coba lagi nanti", nil)
		}
		return c.Next()
	}
}

const (
	registerOTPRateLimitMax    = 5
	registerOTPRateLimitWindow = 10 * time.Minute
)

var registerOTPLimiter = newIPRateLimiter(registerOTPRateLimitMax*devMultiplier(), registerOTPRateLimitWindow)

func RegisterOTPRateLimiter() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if !registerOTPLimiter.allow(c.IP()) {
			return utils.Fail(c, fiber.StatusTooManyRequests,
				"terlalu banyak permintaan kode verifikasi dari alamat ini, coba lagi nanti", nil)
		}
		return c.Next()
	}
}

const (
	passwordResetRateLimitMax    = 5
	passwordResetRateLimitWindow = 10 * time.Minute
)

var passwordResetLimiter = newIPRateLimiter(passwordResetRateLimitMax*devMultiplier(), passwordResetRateLimitWindow)

func PasswordResetRateLimiter() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if !passwordResetLimiter.allow(c.IP()) {
			return utils.Fail(c, fiber.StatusTooManyRequests,
				"terlalu banyak permintaan reset password dari alamat ini, coba lagi nanti", nil)
		}
		return c.Next()
	}
}

const (
	usernameCheckRateLimitMax    = 60
	usernameCheckRateLimitWindow = time.Minute
)

var usernameCheckLimiter = newIPRateLimiter(usernameCheckRateLimitMax*devMultiplier(), usernameCheckRateLimitWindow)

func UsernameCheckRateLimiter() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if !usernameCheckLimiter.allow(c.IP()) {
			return utils.Fail(c, fiber.StatusTooManyRequests,
				"terlalu banyak permintaan, coba lagi sebentar lagi", nil)
		}
		return c.Next()
	}
}

const (
	globalRateLimitMax    = 300
	globalRateLimitWindow = time.Minute
)

var globalLimiter = newIPRateLimiter(globalRateLimitMax, globalRateLimitWindow)

func GlobalRateLimiter() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if !globalLimiter.allow(c.IP()) {
			return utils.Fail(c, fiber.StatusTooManyRequests,
				"terlalu banyak permintaan dari alamat ini, coba lagi sebentar lagi", nil)
		}
		return c.Next()
	}
}
