package auth

import (
	"strconv"
	"sync"
	"time"
)

type totpReplayGuard struct {
	mu   sync.Mutex
	used map[string]time.Time
	ttl  time.Duration
}

func newTOTPReplayGuard(ttl time.Duration) *totpReplayGuard {
	g := &totpReplayGuard{used: make(map[string]time.Time), ttl: ttl}
	go g.cleanupLoop()
	return g
}

func totpReplayKey(userID uint, code string) string {
	return strconv.FormatUint(uint64(userID), 10) + ":" + code
}

// checkAndMark mengembalikan true kalau kode ini BELUM pernah dipakai user
// tsb (dan langsung menandainya sebagai terpakai), atau false kalau ini
// percobaan ulang kode yang sama.
func (g *totpReplayGuard) checkAndMark(userID uint, code string) bool {
	key := totpReplayKey(userID, code)
	g.mu.Lock()
	defer g.mu.Unlock()
	if _, ok := g.used[key]; ok {
		return false
	}
	g.used[key] = time.Now().Add(g.ttl)
	return true
}

func (g *totpReplayGuard) cleanupLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		now := time.Now()
		g.mu.Lock()
		for key, exp := range g.used {
			if now.After(exp) {
				delete(g.used, key)
			}
		}
		g.mu.Unlock()
	}
}
