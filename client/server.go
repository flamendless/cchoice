package client

import (
	"cchoice/client/common"
	"cchoice/client/components"
	"cchoice/client/handlers"
	"cchoice/client/middlewares"
	"cchoice/internal/ctx"
	"cchoice/internal/logs"
	pb "cchoice/proto"
	"embed"
	"encoding/gob"
	"fmt"
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
	"go.uber.org/zap"
)

//go:embed static/*
var static embed.FS

var sessionManager *scs.SessionManager

func Serve(ctxClient *ctx.ClientFlags) {
	addr := fmt.Sprintf("%s%s", ctxClient.Address, ctxClient.Port)
	logs.Log().Info("Starting site server", zap.String("address", addr))

	sessionManager = scs.New()
	sessionManager.Lifetime = 24 * time.Hour
	gob.Register(common.User{})
	gob.Register(common.AuthSession{})

	grpcConn := NewGRPCConn(ctxClient.GRPCAddress, ctxClient.TSC)
	defer GRPCConnectionClose(grpcConn)

	logger := logs.Log()

	mwAuth := middlewares.NewAuthenticated(
		sessionManager,
		pb.NewAuthServiceClient(grpcConn),
		pb.NewUserServiceClient(grpcConn),
	)

	authService := pb.NewAuthServiceClient(grpcConn)
	brandService := pb.NewBrandServiceClient(grpcConn)
	userService := pb.NewUserServiceClient(grpcConn)
	otpService := pb.NewOTPServiceClient(grpcConn)
	shopService := pb.NewShopServiceClient(grpcConn)

	errHandler := handlers.NewErrorHandler(logger)
	authHandler := handlers.NewAuthHandler(logger, authService, sessionManager, mwAuth)
	brandHandler := handlers.NewBrandHandler(logger, brandService)
	userHandler := handlers.NewUserHandler(logger, userService, sessionManager)
	otpHandler := handlers.NewOTPHandler(logger, otpService, authService, sessionManager, mwAuth)
	shopHandler := handlers.NewShopHandler(logger, shopService, sessionManager)

	// productService := services.NewProductService(grpcConn)
	// productHandler := handlers.NewProductHandler(logger, &productService, &authService)

	mux := http.NewServeMux()
	mux.Handle("GET /static/", http.FileServer(http.FS(static)))

	//UTILS-LIKE
	mux.HandleFunc("GET /close_error_banner", func(w http.ResponseWriter, r *http.Request) {
		components.ErrorBanner().Render(r.Context(), w)
	})

	//AUTH
	mux.HandleFunc("GET /auth", errHandler.Default(authHandler.AuthPage))
	mux.HandleFunc("POST /auth", errHandler.Default(authHandler.Authenticate))
	mux.HandleFunc("GET /auth/avatar",  errHandler.Default(authHandler.Avatar))

	//REGISTER
	mux.HandleFunc("GET /register", errHandler.Default(userHandler.RegisterPage))
	mux.HandleFunc("POST /register", errHandler.Default(userHandler.Register))

	//OTP
	mux.HandleFunc("GET /otp", errHandler.Default(otpHandler.OTPView))
	mux.HandleFunc("POST /otp", errHandler.Default(otpHandler.OTPValidate))
	mux.HandleFunc("GET /otp-enroll", errHandler.Default(otpHandler.OTPEnrollView))
	mux.HandleFunc("POST /otp-enroll", errHandler.Default(otpHandler.OTPEnrollFinish))

	//PRODUCTS
	// mux.HandleFunc("GET /products", errHandler.Default(productHandler.ProductTablePage))

	//SHOP
	mux.HandleFunc("GET /", errHandler.Default(shopHandler.HomePage))
	mux.HandleFunc("GET /home", errHandler.Default(shopHandler.HomePage))

	//BRAND
	mux.HandleFunc("GET /brand/{id}", errHandler.Default(brandHandler.BrandPage))
	mux.HandleFunc("GET /brand-logos", errHandler.Default(brandHandler.BrandLogos))

	mw := middlewares.NewMiddleware(
		mux,
		middlewares.WithSessionID(true),
		middlewares.WithSecure(ctxClient.Secure),
		middlewares.WithHTTPOnly(false),
		middlewares.WithGRPC(grpcConn != nil),
		middlewares.WithRequestDurMetrics(true),
	)

	mw = sessionManager.LoadAndSave(mw)

	http.ListenAndServe(addr, mw)
}
