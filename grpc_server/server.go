package grpc_server

import (
	"cchoice/grpc_server/middlewares"
	"cchoice/grpc_server/products"
	"cchoice/internal/ctx"
	"cchoice/internal/logs"
	pb "cchoice/proto"
	"net"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/ratelimit"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func Serve(ctxGRPC ctx.GRPCFlags) {
	lis, err := net.Listen("tcp", ctxGRPC.Address)
	if err != nil {
		logs.Log().Fatal("Failed connection", zap.Error(err))
		return
	}

	logger, opts := middlewares.AddLogger(&ctxGRPC)

	s := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			logging.UnaryServerInterceptor(middlewares.InterceptorLogger(logger), opts...),
			ratelimit.UnaryServerInterceptor(middlewares.AddRateLimit(&ctxGRPC)),
			recovery.UnaryServerInterceptor(middlewares.AddRecovery()...),
		),
		grpc.ChainStreamInterceptor(
			logging.StreamServerInterceptor(middlewares.InterceptorLogger(logger), opts...),
			ratelimit.StreamServerInterceptor(middlewares.AddRateLimit(&ctxGRPC)),
			recovery.StreamServerInterceptor(middlewares.AddRecovery()...),
		),
	)

	ctxDB := ctx.NewDatabaseCtx(ctxGRPC.DBPath)
	defer ctxDB.Close()

	pb.RegisterProductServiceServer(s, &products.ProductServer{CtxDB: ctxDB})
	pb.RegisterProductCategoryServiceServer(s, &products.ProductCategoryServer{CtxDB: ctxDB})
	pb.RegisterProductSpecsServiceServer(s, &products.ProductSpecsServer{CtxDB: ctxDB})

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
