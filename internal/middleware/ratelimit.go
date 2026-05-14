package middleware

import (
	"net/http"
	"sync"
	"time"
)

type rateLimiter struct {
	mu       sync.Mutex
	clients  map[string]*clientInfo
	limit    int
	window   time.Duration
}

type clientInfo struct {
	count   int
	expires time.Time
}

func RateLimitMiddleware(limit int, window time.Duration) func(http.Handler) http.Handler {
	rl := &rateLimiter{
		clients: make(map[string]*clientInfo),
		limit:   limit,
		window:  window,
	}

	go rl.cleanup()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr
			if rl.isLimited(ip) {
				http.Error(w, `{"error":"too many requests"}`, http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func (rl *rateLimiter) isLimited(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	c, ok := rl.clients[key]
	if !ok || now.After(c.expires) {
		rl.clients[key] = &clientInfo{count: 1, expires: now.Add(rl.window)}
		return false
	}
	c.count++
	return c.count > rl.limit
}

func (rl *rateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for k, v := range rl.clients {
			if now.After(v.expires) {
				delete(rl.clients, k)
			}
		}
		rl.mu.Unlock()
	}
}
