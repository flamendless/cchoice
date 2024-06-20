package client

import (
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

func putHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("hi")
	sessionManager.Put(r.Context(), "message", "Hello from a session!")
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	msg := sessionManager.GetString(r.Context(), "message")
	fmt.Println(msg)
}

func Serve(ctxClient *ctx.ClientFlags) {
	addr := fmt.Sprintf("%s%s", ctxClient.Address, ctxClient.Port)
	logs.Log().Info("Starting site server", zap.String("address", addr))

	sessionManager = scs.New()
	sessionManager.Lifetime = 24 * time.Hour

	grpcConn := NewGRPCConn(ctxClient.GRPCAddress)
	defer GRPCConnectionClose(grpcConn)

	logger := logs.Log()

	productService := services.NewProductService(grpcConn)
	productHandler := handlers.NewProductHandler(logger, &productService)

	mux := http.NewServeMux()

	mux.Handle("GET /static/", http.FileServer(http.FS(static)))

	// mux.HandleFunc("GET /", getHandler)
	// mux.HandleFunc("PUT /", putHandler)

	mux.HandleFunc("GET /products", productHandler.ProductTablePage)
	mux.HandleFunc("GET /products_table", productHandler.ProductTableBody)

	mw := middlewares.NewMiddleware(
		mux,
		middlewares.WithSecure(ctxClient.Secure),
		middlewares.WithGRPC(ctxClient.GRPCAddress != ""),
	)

	mw = sessionManager.LoadAndSave(mw)

	http.ListenAndServe(addr, mw)
}
