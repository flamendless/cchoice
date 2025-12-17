package server

import (
	"net/http"
)

func SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("X-Download-Options", "noopen")
		csp := "default-src 'self'; " +
			"script-src 'self' 'unsafe-eval' 'unsafe-inline' https://unpkg.com ajax.cloudflare.com static.cloudflareinsights.com; " +
			"style-src 'self' 'unsafe-inline'; " +
			"img-src 'self' data: https:; " +
			"connect-src 'self' cloudflareinsights.com; " +
			"font-src 'self' data: fonts.googleapis.com fonts.gstatic.com; " +
			"object-src 'none'; " +
			"media-src 'self'; " +
			"frame-ancestors 'none'; " +
			"base-uri 'self'; " +
			"form-action 'self'; " +
			"upgrade-insecure-requests;"
		w.Header().Set("Content-Security-Policy", csp)

		permissions := "accelerometer=(), " +
			"camera=(), " +
			"geolocation=(), " +
			"gyroscope=(), " +
			"magnetometer=(), " +
			"microphone=(), " +
			"payment=(), " +
			"usb=()"
		w.Header().Set("Permissions-Policy", permissions)

		next.ServeHTTP(w, r)
	})
}

func RateLimitInfoMiddleware(limit int, remaining int, reset int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Add rate limit headers for API transparency
			// w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
			// w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
			// w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", reset))

			next.ServeHTTP(w, r)
		})
	}
}
