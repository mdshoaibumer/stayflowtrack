package middleware

import (
	"net/http"
	"sync"
	"time"

	apperrors "github.com/stayflow/stayflow-track/internal/shared/errors"
	"github.com/stayflow/stayflow-track/internal/shared/response"
)

// RateLimiter implements a sliding window rate limiter per IP.
type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	limit    int
	window   time.Duration
}

type visitor struct {
	timestamps []time.Time
}

// NewRateLimiter creates a rate limiter allowing `limit` requests per `window` per IP.
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		limit:    limit,
		window:   window,
	}
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.window)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, v := range rl.visitors {
			v.timestamps = filterRecent(v.timestamps, now, rl.window)
			if len(v.timestamps) == 0 {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

func filterRecent(timestamps []time.Time, now time.Time, window time.Duration) []time.Time {
	cutoff := now.Add(-window)
	var recent []time.Time
	for _, t := range timestamps {
		if t.After(cutoff) {
			recent = append(recent, t)
		}
	}
	return recent
}

func (rl *RateLimiter) isAllowed(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	v, exists := rl.visitors[ip]
	if !exists {
		rl.visitors[ip] = &visitor{timestamps: []time.Time{now}}
		return true
	}

	v.timestamps = filterRecent(v.timestamps, now, rl.window)
	if len(v.timestamps) >= rl.limit {
		return false
	}

	v.timestamps = append(v.timestamps, now)
	return true
}

// Limit returns middleware that rate limits requests by client IP.
func (rl *RateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
			ip = fwd
		}

		if !rl.isAllowed(ip) {
			w.Header().Set("Retry-After", "60")
			response.Err(w, apperrors.New(http.StatusTooManyRequests, "RATE_LIMITED", "too many requests, please try again later"))
			return
		}

		next.ServeHTTP(w, r)
	})
}
