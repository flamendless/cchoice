package client

import (
	"cchoice/client/components"
	"cchoice/client/handlers"
	"cchoice/client/middlewares"
	"cchoice/client/services"
	"cchoice/internal/ctx"
	"cchoice/internal/logs"
	"embed"
	"fmt"
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
	"go.uber.org/zap"
)

//go:embed static/*
var static embed.FS

var sessionManager *scs.SessionManager

// func putHandler(w http.ResponseWriter, r *http.Request) {
// 	fmt.Println("hi")
// 	sessionManager.Put(r.Context(), "message", "Hello from a session!")
// }
//
// func getHandler(w http.ResponseWriter, r *http.Request) {
// 	msg := sessionManager.GetString(r.Context(), "message")
// 	fmt.Println(msg)
// }

func Serve(ctxClient *ctx.ClientFlags) {
	addr := fmt.Sprintf("%s%s", ctxClient.Address, ctxClient.Port)
	logs.Log().Info("Starting site server", zap.String("address", addr))

	sessionManager = scs.New()
	sessionManager.Lifetime = 24 * time.Hour

	grpcConn := NewGRPCConn(ctxClient.GRPCAddress, ctxClient.TSC)
	defer GRPCConnectionClose(grpcConn)

	logger := logs.Log()

	errHandler := handlers.NewErrorHandler(logger)

	productService := services.NewProductService(grpcConn)
	productHandler := handlers.NewProductHandler(logger, &productService)

	authService := services.NewAuthService(grpcConn)
	authHandler := handlers.NewAuthHandler(logger, &authService, sessionManager)

	mux := http.NewServeMux()

	mux.Handle("GET /static/", http.FileServer(http.FS(static)))

	//UTILS-LIKe
	mux.HandleFunc("GET /close_error_banner", func(w http.ResponseWriter, r *http.Request) {
		components.ErrorBanner().Render(r.Context(), w)
	})

	//AUTH
	mux.HandleFunc("GET /", errHandler.Default(authHandler.AuthPage))
	mux.HandleFunc("GET /auth", errHandler.Default(authHandler.AuthPage))
	mux.HandleFunc("POST /auth", errHandler.Default(authHandler.Authenticate))

	//PRODUCTS
	mux.HandleFunc("GET /products", errHandler.Default(productHandler.ProductTablePage))

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
