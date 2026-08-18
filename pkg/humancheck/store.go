package humancheck

import (
	"sync"
	"time"
)

type usedStore struct {
	mu   sync.Mutex
	data map[string]time.Time // token -> waktu kedaluwarsa entry ini boleh dibuang
}

func newUsedStore() *usedStore {
	s := &usedStore{data: make(map[string]time.Time)}
	go s.cleanupLoop()
	return s
}

func (s *usedStore) isUsed(token string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.data[token]
	return ok
}

func (s *usedStore) markUsed(token string, ttl time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[token] = time.Now().Add(ttl)
}

func (s *usedStore) cleanupLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		now := time.Now()
		s.mu.Lock()
		for token, exp := range s.data {
			if now.After(exp) {
				delete(s.data, token)
			}
		}
		s.mu.Unlock()
	}
}
