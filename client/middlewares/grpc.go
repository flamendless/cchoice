package middlewares

func WithGRPC(useGRPC bool) MiddlewareOpts {
	return func(m *Middleware) {
		m.UseGRPC = useGRPC
	}
}
