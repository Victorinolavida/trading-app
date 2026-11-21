// Package passwordreset contains safeguards for password reset requests.
package passwordreset

import (
	"crypto/sha256"
	"strings"
	"sync"
	"time"
)

type entry struct {
	count   int
	expires time.Time
}

// Limiter applies fixed-window limits per IP address and email address.
// It is process-local; deployments with multiple replicas need a shared store.
type Limiter struct {
	mu      sync.Mutex
	entries map[[32]byte]entry
	now     func() time.Time
}

func NewLimiter() *Limiter {
	return &Limiter{entries: make(map[[32]byte]entry), now: time.Now}
}

// Allow limits requests to 10 per IP and 3 per address per 15 minutes.
// The caller must provide an IP obtained from a trusted source.
func (l *Limiter) Allow(ip, email string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	for key, value := range l.entries {
		if !now.Before(value.expires) {
			delete(l.entries, key)
		}
	}
	ipKey := sha256.Sum256([]byte("ip:" + ip))
	emailKey := sha256.Sum256([]byte("email:" + strings.ToLower(strings.TrimSpace(email))))
	if l.entries[ipKey].count >= 10 || l.entries[emailKey].count >= 3 {
		return false
	}
	// Bound memory usage under a flood of distinct addresses.
	if len(l.entries) >= 100000 {
		return false
	}
	for _, key := range [][32]byte{ipKey, emailKey} {
		value := l.entries[key]
		if value.count == 0 {
			value.expires = now.Add(15 * time.Minute)
		}
		value.count++
		l.entries[key] = value
	}
	return true
}
