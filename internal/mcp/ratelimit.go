package mcp

import (
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimiter provides simple per-IP token bucket rate limiting.
// For a production system this would be backed by Redis or a distributed store.
type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.Mutex
	rps      float64
	burst    int
}

// NewRateLimiter creates a new rate limiter.
func NewRateLimiter(rps float64, burst int) *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rps:      rps,
		burst:    burst,
	}
}

func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.limiters[ip]
	if !exists {
		limiter = rate.NewLimiter(rate.Limit(rl.rps), rl.burst)
		rl.limiters[ip] = limiter

		// Simple cleanup: remove old limiters after some time (basic memory hygiene)
		go func(key string) {
			time.Sleep(10 * time.Minute)
			rl.mu.Lock()
			delete(rl.limiters, key)
			rl.mu.Unlock()
		}(ip)
	}
	return limiter
}

// Middleware returns an http.Handler middleware that enforces rate limits.
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			ip = xff // best-effort for reverse proxies
		}

		limiter := rl.getLimiter(ip)
		if !limiter.Allow() {
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
