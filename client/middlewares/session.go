package middlewares

import "net/http"

func WithSessionID(useSessionID bool) MiddlewareOpts {
	return func(m *Middleware) {
		m.UseSessionID = useSessionID
	}
}

func SessionID(r *http.Request) string {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		return ""
	}
	return cookie.Value
}
