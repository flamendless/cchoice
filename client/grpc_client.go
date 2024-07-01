package client

import (
	"cchoice/conf"
	"cchoice/internal/auth"
	"cchoice/internal/enums"
	"cchoice/internal/logs"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewGRPCConn(addr string, tsc bool) *grpc.ClientConn {
	issuer, err := auth.NewIssuer()
	if err != nil {
		panic(err)
	}

	tokenString, err := issuer.IssueToken(
		enums.AudAPI,
		conf.GetConf().ClientUsername,
	)
	if err != nil {
		panic(err)
	}

	var opts []grpc.DialOption
	opts = append(
		opts,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithPerRPCCredentials(auth.ClientToken{
			Token: tokenString,
			TSC:   tsc,
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
