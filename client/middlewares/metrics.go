package middlewares

func WithRequestDurMetrics(flag bool) MiddlewareOpts {
	return func(m *Middleware) {
		m.RequestDurMetrics = flag
	}
}
