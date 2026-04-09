package middleware

import (
	"net/http"
	"sync"
	"time"
)

type rateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	rate     int
	window   time.Duration
}

type visitor struct {
	count    int
	resetAt  time.Time
}

// RateLimit returns a middleware that limits requests per IP
func RateLimit(requestsPerMinute int) func(http.Handler) http.Handler {
	rl := &rateLimiter{
		visitors: make(map[string]*visitor),
		rate:     requestsPerMinute,
		window:   time.Minute,
	}

	// Clean up stale entries periodically
	go func() {
		for {
			time.Sleep(time.Minute)
			rl.mu.Lock()
			now := time.Now()
			for ip, v := range rl.visitors {
				if now.After(v.resetAt) {
					delete(rl.visitors, ip)
				}
			}
			rl.mu.Unlock()
		}
	}()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr

			rl.mu.Lock()
			v, exists := rl.visitors[ip]
			now := time.Now()

			if !exists || now.After(v.resetAt) {
				rl.visitors[ip] = &visitor{
					count:   1,
					resetAt: now.Add(rl.window),
				}
				rl.mu.Unlock()
				next.ServeHTTP(w, r)
				return
			}

			v.count++
			if v.count > rl.rate {
				rl.mu.Unlock()
				w.Header().Set("Retry-After", "60")
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			rl.mu.Unlock()

			next.ServeHTTP(w, r)
		})
	}
}
