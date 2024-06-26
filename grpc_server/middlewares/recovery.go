package middlewares

import (
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func fnRecovery(p any) error {
	return status.Errorf(codes.Unknown, "panic triggered: %v", p)
}

func AddRecovery() []recovery.Option {
	return []recovery.Option{
		recovery.WithRecoveryHandler(fnRecovery),
	}
}
