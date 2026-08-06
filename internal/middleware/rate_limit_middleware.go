package middleware

import (
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/projsonal/gowms/pkg/utils"
)

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

var registerLimiter = newIPRateLimiter(registerRateLimitMax, registerRateLimitWindow)

func RegisterRateLimiter() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if !registerLimiter.allow(c.IP()) {
			return utils.Fail(c, fiber.StatusTooManyRequests,
				"terlalu banyak percobaan pendaftaran dari alamat ini, coba lagi nanti", nil)
		}
		return c.Next()
	}
}
