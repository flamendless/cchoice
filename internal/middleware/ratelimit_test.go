package middleware

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"golang.org/x/time/rate"
)

func TestAllowsUnderLimit(t *testing.T) {
	rl := NewRateLimiterWithDebug(rate.Limit(5), 10, time.Minute, false)
	defer rl.Stop()

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	})

	middleware := rl.Middleware(next)

	req := httptest.NewRequest(http.MethodPost, "/customer/login", nil)
	req.RemoteAddr = "192.168.1.1:1234"

	for i := 0; i < 5; i++ {
		w := httptest.NewRecorder()
		middleware.ServeHTTP(w, req)
		if !nextCalled {
			t.Fatalf("expected next handler to be called on request %d", i+1)
		}
		nextCalled = false
	}
}

func TestBlocksOverLimit(t *testing.T) {
	rl := NewRateLimiterWithDebug(rate.Limit(2), 2, time.Minute, false)
	defer rl.Stop()

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	})

	middleware := rl.Middleware(next)

	req := httptest.NewRequest(http.MethodPost, "/customer/login", nil)
	req.RemoteAddr = "192.168.1.1:1234"

	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		middleware.ServeHTTP(w, req)
		if !nextCalled {
			t.Fatalf("expected next handler to be called on request %d", i+1)
		}
		nextCalled = false
	}

	w := httptest.NewRecorder()
	middleware.ServeHTTP(w, req)

	if nextCalled {
		t.Fatal("expected next handler to NOT be called when rate limit exceeded")
	}

	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("expected status %d, got %d", http.StatusTooManyRequests, w.Code)
	}

	if w.Header().Get("Retry-After") != "30" {
		t.Fatalf("expected Retry-After header to be 30, got %s", w.Header().Get("Retry-After"))
	}
}

func TestCleanupExpiredVisitors(t *testing.T) {
	rl := NewRateLimiterWithDebug(rate.Limit(5), 10, 100*time.Millisecond, false)
	defer rl.Stop()

	req := httptest.NewRequest(http.MethodPost, "/customer/login", nil)
	req.RemoteAddr = "192.168.1.1:1234"

	rl.getVisitor(rl.getIP(req))

	time.Sleep(150 * time.Millisecond)

	rl.cleanup()

	rl.mu.Lock()
	if len(rl.visitors) != 0 {
		t.Fatalf("expected 0 visitors after cleanup, got %d", len(rl.visitors))
	}
	rl.mu.Unlock()
}

func TestConcurrentAccess(t *testing.T) {
	rl := NewRateLimiterWithDebug(rate.Limit(100), 100, time.Minute, false)
	defer rl.Stop()

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	middleware := rl.Middleware(next)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				req := httptest.NewRequest(http.MethodPost, "/customer/login", nil)
				req.RemoteAddr = "192.168.1.1:1234"
				w := httptest.NewRecorder()
				middleware.ServeHTTP(w, req)
			}
		}()
	}

	wg.Wait()
}

func TestIPExtraction(t *testing.T) {
	rl := NewRateLimiterWithDebug(rate.Limit(5), 10, time.Minute, false)
	defer rl.Stop()

	tests := []struct {
		name         string
		remoteAddr   string
		forwardedFor string
		realIP       string
		expectedIP   string
	}{
		{
			name:       "RemoteAddr only",
			remoteAddr: "192.168.1.1:1234",
			expectedIP: "192.168.1.1",
		},
		{
			name:         "X-Forwarded-For",
			remoteAddr:   "192.168.1.1:1234",
			forwardedFor: "10.0.0.1, 192.168.1.1",
			expectedIP:   "10.0.0.1",
		},
		{
			name:       "X-Real-IP",
			remoteAddr: "192.168.1.1:1234",
			realIP:     "10.0.0.2",
			expectedIP: "10.0.0.2",
		},
		{
			name:         "X-Forwarded-For takes priority over X-Real-IP",
			remoteAddr:   "192.168.1.1:1234",
			forwardedFor: "10.0.0.1",
			realIP:       "10.0.0.2",
			expectedIP:   "10.0.0.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/customer/login", nil)
			req.RemoteAddr = tt.remoteAddr
			if tt.forwardedFor != "" {
				req.Header.Set("X-Forwarded-For", tt.forwardedFor)
			}
			if tt.realIP != "" {
				req.Header.Set("X-Real-IP", tt.realIP)
			}

			ip := rl.getIP(req)
			if ip != tt.expectedIP {
				t.Errorf("expected IP %s, got %s", tt.expectedIP, ip)
			}
		})
	}
}

func TestDifferentIPs(t *testing.T) {
	rl := NewRateLimiterWithDebug(rate.Limit(1), 1, time.Minute, false)
	defer rl.Stop()

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	})

	middleware := rl.Middleware(next)

	req1 := httptest.NewRequest(http.MethodPost, "/customer/login", nil)
	req1.RemoteAddr = "192.168.1.1:1234"

	w := httptest.NewRecorder()
	middleware.ServeHTTP(w, req1)
	if !nextCalled {
		t.Fatal("expected first request to succeed")
	}
	nextCalled = false

	req2 := httptest.NewRequest(http.MethodPost, "/customer/login", nil)
	req2.RemoteAddr = "192.168.1.2:1234"

	w = httptest.NewRecorder()
	middleware.ServeHTTP(w, req2)
	if !nextCalled {
		t.Fatal("expected request from different IP to succeed")
	}
}
