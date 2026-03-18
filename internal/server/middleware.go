package server

import (
	"cchoice/internal/conf"
	"cchoice/internal/enums"
	"cchoice/internal/metrics"
	"cchoice/internal/utils"
	"context"
	"crypto/subtle"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5/middleware"
	"golang.org/x/crypto/bcrypt"
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
			"geolocation=(self), " +
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

func MetricsBasicAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok {
			unauthorized(w)
			return
		}

		cfg := conf.Conf()
		expectedUser := cfg.BasicAuth.Username
		passHash := cfg.BasicAuth.PasswordHash

		if subtle.ConstantTimeCompare([]byte(user), []byte(expectedUser)) != 1 {
			unauthorized(w)
			return
		}

		if bcrypt.CompareHashAndPassword([]byte(passHash), []byte(pass)) != nil {
			unauthorized(w)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func unauthorized(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Basic realm="metrics"`)
	w.WriteHeader(http.StatusUnauthorized)
}

func PrometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		if utils.MatchPath(path, "/metrics") || utils.MatchPath(path, "/health") {
			metrics.HTTP.RoutesSkippedHit(r.Method, path)
			next.ServeHTTP(w, r)
			return
		}

		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		next.ServeHTTP(ww, r)

		status := ww.Status()

		metrics.HTTP.RequestsHit(r.Method, path, strconv.Itoa(status))

		if status >= 500 {
			metrics.HTTP.ErrorsHit(r.Method, path, strconv.Itoa(status))
		}
	})
}

func (s *Server) requireStaffAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		staffID := s.encoder.Decode(s.sessionManager.GetString(r.Context(), SessionStaffID))
		if staffID == 0 {
			http.Redirect(w, r, utils.URL("/admin"), http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) requireSuperuserAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		staff, err := s.services.staff.GetCurrentStaff(r.Context(), s.sessionManager.GetString(r.Context(), SessionStaffID))
		if staff.ID == 0 {
			http.Redirect(w, r, utils.URL("/admin"), http.StatusSeeOther)
			return
		}

		if err != nil || staff.UserType != enums.STAFF_USER_TYPE_SUPERUSER.String() {
			http.Redirect(w, r, utils.URL("/admin/staff"), http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) HasRole(ctx context.Context, role enums.StaffRole) bool {
	staffIDStr := s.sessionManager.GetString(ctx, SessionStaffID)
	if staffIDStr == "" {
		return false
	}

	staffID := s.encoder.Decode(staffIDStr)
	if staffID == 0 {
		return false
	}

	staff, err := s.services.staff.GetCurrentStaff(ctx, staffIDStr)
	if err != nil {
		return false
	}

	if staff.UserType == enums.STAFF_USER_TYPE_SUPERUSER.String() {
		return true
	}

	roles, err := s.dbRO.GetQueries().GetStaffRolesByStaffID(ctx, staffID)
	if err != nil {
		return false
	}

	roleStr := role.String()
	for _, r := range roles {
		if r == roleStr {
			return true
		}
	}
	return false
}

func (s *Server) AllowRoles(roles ...enums.StaffRole) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			hasAccess := false

			for _, role := range roles {
				if s.HasRole(ctx, role) {
					hasAccess = true
					break
				}
			}

			if !hasAccess {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
