package middleware

import (
	"net/http"
	"sync"
	"time"
)

// LoginRateLimiter is a simple fixed-window limiter keyed by client IP,
// closing the legacy gap where /proseslogin and /prosesloginadmin had no
// throttling at all (A4/D-4). In-memory and single-instance only -- if the
// API is ever scaled horizontally this needs to move to a shared store
// (e.g. Redis) instead; tracked as a follow-up, not silently ignored.
type LoginRateLimiter struct {
	mu       sync.Mutex
	window   time.Duration
	limit    int
	attempts map[string][]time.Time
}

func NewLoginRateLimiter(limit int, window time.Duration) *LoginRateLimiter {
	return &LoginRateLimiter{
		window:   window,
		limit:    limit,
		attempts: make(map[string][]time.Time),
	}
}

func (l *LoginRateLimiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-l.window)

	kept := l.attempts[key][:0]
	for _, t := range l.attempts[key] {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}

	if len(kept) >= l.limit {
		l.attempts[key] = kept
		return false
	}

	l.attempts[key] = append(kept, now)
	return true
}

func (l *LoginRateLimiter) Middleware(keyFn func(r *http.Request) string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !l.Allow(keyFn(r)) {
				http.Error(w, "too many login attempts, try again later", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
