package middlewares

import (
	"cchoice/internal/logs"
	"net/http"

	"github.com/segmentio/ksuid"
	"go.uber.org/zap"
)

type Middleware struct {
	Next         http.Handler
	UseSessionID bool
	Secure       bool
	HTTPOnly     bool
	UseGRPC      bool
}

type MiddlewareOpts func(*Middleware)

func NewMiddleware(next http.Handler, opts ...MiddlewareOpts) http.Handler {
	mw := Middleware{
		Next:         next,
		UseSessionID: true,
		Secure:       true,
		HTTPOnly:     false,
		UseGRPC:      false,
	}
	for _, opt := range opts {
		opt(&mw)
	}

	logs.Log().Info(
		"Middlewares",
		zap.Bool("session ID", mw.UseSessionID),
		zap.Bool("secure", mw.Secure),
		zap.Bool("HTTP only", mw.HTTPOnly),
		zap.Bool("GRPC", mw.UseGRPC),
	)

	return mw
}

func (mw Middleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if mw.UseSessionID {
		id := SessionID(r)
		if id == "" {
			id = ksuid.New().String()
			http.SetCookie(w, &http.Cookie{
				Name:     "session_id",
				Value:    id,
				Secure:   mw.Secure,
				HttpOnly: mw.HTTPOnly,
			})
		}
	}

	mw.Next.ServeHTTP(w, r)
}
