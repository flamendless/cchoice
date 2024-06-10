package grpc_server

import (
	"cchoice/internal/ctx"
	"cchoice/internal/logs"
	pb "cchoice/proto"
	"net"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func Serve(ctxGRPC ctx.GRPCFlags) {
	lis, err := net.Listen("tcp", ctxGRPC.Port)
	if err != nil {
		logs.Log().Fatal("Failed connection", zap.Error(err))
		return
	}

	s := grpc.NewServer()

	ctxDB := ctx.NewDatabaseCtx(ctxGRPC.DBPath)
	defer ctxDB.Close()

	pb.RegisterProductServiceServer(s, &ProductServer{ctxDB: ctxDB})

	if ctxGRPC.Reflection {
		reflection.Register(s)
	}

	logs.Log().Info(
		"Server",
		zap.String("address", lis.Addr().String()),
		zap.String("network", lis.Addr().Network()),
	)

	err = s.Serve(lis)
	if err != nil {
		logs.Log().Fatal("Server failed", zap.Error(err))
	}
}
