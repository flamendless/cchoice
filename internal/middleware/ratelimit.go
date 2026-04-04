package middleware

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"cchoice/internal/logs"
	"cchoice/internal/metrics"

	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

type Visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type RateLimiter struct {
	visitors map[string]*Visitor
	mu       sync.Mutex

	rate  rate.Limit
	burst int
	ttl   time.Duration
	debug bool

	stopCleanup chan struct{}
}

func NewRateLimiter(r rate.Limit, burst int, ttl time.Duration) *RateLimiter {
	return NewRateLimiterWithDebug(r, burst, ttl, false)
}

func NewRateLimiterWithDebug(r rate.Limit, burst int, ttl time.Duration, debug bool) *RateLimiter {
	rl := &RateLimiter{
		visitors:    make(map[string]*Visitor),
		rate:        r,
		burst:       burst,
		ttl:         ttl,
		debug:       debug,
		stopCleanup: make(chan struct{}),
	}

	go rl.cleanupLoop()

	return rl
}

func (rl *RateLimiter) Stop() {
	close(rl.stopCleanup)
}

func (rl *RateLimiter) getIP(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		parts := strings.Split(forwarded, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}

	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return strings.TrimSpace(realIP)
	}

	ip := r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}

	return ip
}

func (rl *RateLimiter) getVisitor(ip string) *Visitor {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]
	if !exists {
		limiter := rate.NewLimiter(rl.rate, rl.burst)
		v = &Visitor{
			limiter:  limiter,
			lastSeen: time.Now(),
		}
		rl.visitors[ip] = v
	} else {
		v.lastSeen = time.Now()
	}

	return v
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := rl.getIP(r)
		path := r.URL.Path

		if rl.debug {
			logs.Log().Info("[RateLimit] processing request", zap.String("ip", ip), zap.String("path", path))
		}

		visitor := rl.getVisitor(ip)

		if !visitor.limiter.Allow() {
			metrics.HTTP.RateLimitedHit(path, ip)

			if rl.debug {
				logs.Log().Info("[RateLimit] blocked", zap.String("ip", ip), zap.String("path", path))
			}

			w.Header().Set("Retry-After", "30")
			w.WriteHeader(http.StatusTooManyRequests)
			http.Error(w, "Rate limit exceeded. Please try again later.", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-rl.stopCleanup:
			return
		case <-ticker.C:
			rl.cleanup()
		}
	}
}

func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for ip, v := range rl.visitors {
		if now.Sub(v.lastSeen) > rl.ttl {
			delete(rl.visitors, ip)
		}
	}

	metrics.HTTP.SetRateLimitActiveVisitors(float64(len(rl.visitors)))
}
