package grpc_server

import (
	grpcauth "cchoice/grpc_server/auth"
	grpcbrand "cchoice/grpc_server/brand"
	"cchoice/grpc_server/middlewares"
	grpcotp "cchoice/grpc_server/otp"
	"cchoice/grpc_server/products"
	grpcsettings "cchoice/grpc_server/settings"
	grpcshop "cchoice/grpc_server/shop"
	grpcuser "cchoice/grpc_server/user"
	cchoiceauth "cchoice/internal/auth"
	internalauth "cchoice/internal/auth"
	"cchoice/internal/ctx"
	"cchoice/internal/logs"
	pb "cchoice/proto"
	"net"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
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

	ctxDB := ctx.NewDatabaseCtx(ctxGRPC.DBPath)
	defer ctxDB.Close()

	validator, err := cchoiceauth.NewValidator(ctxDB)
	if err != nil {
		panic(err)
	}
	authMW := middlewares.AddAuth(validator)

	s := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			logging.UnaryServerInterceptor(middlewares.InterceptorLogger(logger), opts...),
			ratelimit.UnaryServerInterceptor(middlewares.AddRateLimit(&ctxGRPC)),
			auth.UnaryServerInterceptor(authMW.Handle),
			recovery.UnaryServerInterceptor(middlewares.AddRecovery()...),
		),
		grpc.ChainStreamInterceptor(
			logging.StreamServerInterceptor(middlewares.InterceptorLogger(logger), opts...),
			ratelimit.StreamServerInterceptor(middlewares.AddRateLimit(&ctxGRPC)),
			auth.StreamServerInterceptor(authMW.Handle),
			recovery.StreamServerInterceptor(middlewares.AddRecovery()...),
		),
	)

	issuer, err := internalauth.NewIssuer()
	if err != nil {
		panic(err)
	}

	grpcAuthServer := grpcauth.NewGRPCAuthServer(ctxDB, issuer, validator)
	grpcBrandServer := grpcbrand.NewGRPCBrandServer(ctxDB)
	grpcOTPServer := grpcotp.NewGRPCOTPServer(ctxDB, issuer, validator)
	grpcSettingsServer := grpcsettings.NewGRPCSettingsServer(ctxDB)
	grpcShopServer := grpcshop.NewGRPCShopServer(ctxDB)
	grpcUserServer := grpcuser.NewGRPCUserServer(ctxDB)
	pb.RegisterAuthServiceServer(s, grpcAuthServer)
	pb.RegisterBrandServiceServer(s, grpcBrandServer)
	pb.RegisterOTPServiceServer(s, grpcOTPServer)
	pb.RegisterSettingsServiceServer(s, grpcSettingsServer)
	pb.RegisterShopServiceServer(s, grpcShopServer)
	pb.RegisterUserServiceServer(s, grpcUserServer)
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
