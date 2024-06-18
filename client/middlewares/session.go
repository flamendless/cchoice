package middlewares

import (
	"cchoice/internal/logs"
	"net/http"

	"github.com/segmentio/ksuid"
	"go.uber.org/zap"
)

type Middleware struct {
	Next     http.Handler
	Secure   bool
	HTTPOnly bool
	UseGRPC  bool
}

type MiddlewareOpts func(*Middleware)

func NewMiddleware(next http.Handler, opts ...MiddlewareOpts) http.Handler {
	mw := Middleware{
		Next:     next,
		Secure:   true,
		HTTPOnly: true,
	}
	for _, opt := range opts {
		opt(&mw)
	}
	return mw
}

func WithSecure(secure bool) MiddlewareOpts {
	logs.Log().Info("Middleware", zap.Bool("with secure", secure))
	return func(m *Middleware) {
		m.Secure = secure
	}
}

func WithHTTPOnly(httpOnly bool) MiddlewareOpts {
	logs.Log().Info("Middleware", zap.Bool("with HTTP only", httpOnly))
	return func(m *Middleware) {
		m.HTTPOnly = httpOnly
	}
}

func WithGRPC(useGRPC bool) MiddlewareOpts {
	logs.Log().Info("Middleware", zap.Bool("use GRPC", useGRPC))
	return func(m *Middleware) {
		m.UseGRPC = useGRPC
	}
}

func ID(r *http.Request) string {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		return ""
	}
	return cookie.Value
}

func (mw Middleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id := ID(r)
	if id == "" {
		id = ksuid.New().String()
		http.SetCookie(w, &http.Cookie{
			Name:     "session_id",
			Value:    id,
			Secure:   mw.Secure,
			HttpOnly: mw.HTTPOnly,
		})
	}

	mw.Next.ServeHTTP(w, r)
}
