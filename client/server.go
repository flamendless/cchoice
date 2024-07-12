package client

import (
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
	gob.Register(pb.RegisterResponse{})

	grpcConn := NewGRPCConn(ctxClient.GRPCAddress, ctxClient.TSC)
	defer GRPCConnectionClose(grpcConn)

	logger := logs.Log()

	errHandler := handlers.NewErrorHandler(logger)

	authService := pb.NewAuthServiceClient(grpcConn)
	// authService := services.NewAuthService(grpcConn, sessionManager)
	authHandler := handlers.NewAuthHandler(logger, authService, sessionManager)

	userService := pb.NewUserServiceClient(grpcConn)
	userHandler := handlers.NewUserHandler(logger, userService, sessionManager)

	otpService := pb.NewOTPServiceClient(grpcConn)
	otpHandler := handlers.NewOTPHandler(logger, otpService, sessionManager)

	// productService := services.NewProductService(grpcConn)
	// productHandler := handlers.NewProductHandler(logger, &productService, &authService)

	mux := http.NewServeMux()
	mux.Handle("GET /static/", http.FileServer(http.FS(static)))

	//UTILS-LIKE
	mux.HandleFunc("GET /close_error_banner", func(w http.ResponseWriter, r *http.Request) {
		components.ErrorBanner().Render(r.Context(), w)
	})

	//AUTH
	// mux.HandleFunc("GET /", errHandler.Default(userHandler.AuthPage))
	mux.HandleFunc("GET /register", errHandler.Default(userHandler.RegisterPage))
	mux.HandleFunc("POST /register", errHandler.Default(userHandler.Register))

	//REGISTER
	mux.HandleFunc("GET /auth", errHandler.Default(authHandler.AuthPage))
	mux.HandleFunc("POST /auth", errHandler.Default(authHandler.Authenticate))

	//OTP
	// mux.HandleFunc("GET /otp", errHandler.Default(otpHandler.OTPView))
	// mux.HandleFunc("POST /otp", errHandler.Default(otpHandler.OTPValidate))
	mux.HandleFunc("GET /otp-enroll", errHandler.Default(otpHandler.OTPEnrollView))
	mux.HandleFunc("POST /otp-enroll", errHandler.Default(otpHandler.OTPEnrollFinish))

	//PRODUCTS
	// mux.HandleFunc("GET /products", errHandler.Default(productHandler.ProductTablePage))

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
