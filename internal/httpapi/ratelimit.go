package httpapi

import (
	"net"
	"net/http"
	"sync"
	"time"
)

const (
	loginRateLimit  = 5
	loginRateWindow = time.Minute
)

// ipRateLimiter caps how many attempts a single IP gets within a rolling
// window. A plain in-memory map is enough for a two-user deployment.
type ipRateLimiter struct {
	mu       sync.Mutex
	attempts map[string][]time.Time
}

func newIPRateLimiter() *ipRateLimiter {
	return &ipRateLimiter{attempts: make(map[string][]time.Time)}
}

// Allow reports whether ip has not yet exhausted its quota, and records the
// attempt if so.
func (l *ipRateLimiter) Allow(ip string) bool {
	now := time.Now()
	cutoff := now.Add(-loginRateWindow)

	l.mu.Lock()
	defer l.mu.Unlock()

	kept := l.attempts[ip][:0]
	for _, t := range l.attempts[ip] {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}

	if len(kept) >= loginRateLimit {
		l.attempts[ip] = kept
		return false
	}

	l.attempts[ip] = append(kept, now)
	return true
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
