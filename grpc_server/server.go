package grpc_server

import (
	"cchoice/grpc_server/products"
	"cchoice/internal/ctx"
	"cchoice/internal/logs"
	pb "cchoice/proto"
	"net"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
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

	logger := logs.Log()
	var opts []logging.Option

	if ctxGRPC.LogPayloadReceived {
		opts = []logging.Option{
			logging.WithLogOnEvents(
				logging.StartCall,
				logging.FinishCall,
				logging.PayloadReceived,
			),
		}
	} else {
		opts = []logging.Option{
			logging.WithLogOnEvents(
				logging.StartCall,
				logging.FinishCall,
			),
		}
	}

	s := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			logging.UnaryServerInterceptor(logs.InterceptorLogger(logger), opts...),
		),
		grpc.ChainStreamInterceptor(
			logging.StreamServerInterceptor(logs.InterceptorLogger(logger), opts...),
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
