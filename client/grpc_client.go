package client

import (
	"cchoice/internal/auth"
	"cchoice/internal/logs"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewGRPCConn(addr string, tsc bool) *grpc.ClientConn {
	var opts []grpc.DialOption
	opts = append(
		opts,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithPerRPCCredentials(auth.AuthToken{
			Token: "client",
			TSC: tsc,
		}),
	)

	conn, err := grpc.NewClient(addr, opts...)
	if err != nil {
		logs.Log().Fatal("GRPC client", zap.Error(err))
	}
	return conn
}

func GRPCConnectionClose(conn *grpc.ClientConn) {
	defer conn.Close()
}
