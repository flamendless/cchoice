package client

import (
	"cchoice/client/components"
	"cchoice/client/components/layout"
	"cchoice/client/middlewares"
	"cchoice/internal/ctx"
	"cchoice/internal/logs"
	pb "cchoice/proto"
	"context"
	"embed"
	"fmt"
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
	"go.uber.org/zap"
)

//go:embed static/*
var static embed.FS

//go:generate npx tailwindcss build -i static/css/style.css -o static/css/tailwind.css -m

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

	mux := http.NewServeMux()

	mux.Handle("GET /static/", http.FileServer(http.FS(static)))

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		components.Hello("cchoice").Render(r.Context(), w)
	})

	// mux.HandleFunc("GET /", getHandler)
	// mux.HandleFunc("PUT /", putHandler)

	mux.HandleFunc("GET /products", func(w http.ResponseWriter, r *http.Request) {
		//TODO: This should be in a handler?
		client := pb.NewProductServiceClient(grpcConn)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		products, err := client.ListProductsByProductStatus(
			ctx,
			&pb.ProductStatusRequest{Status: pb.ProductStatus_ACTIVE},
		)
		if err != nil {
			logs.Log().Fatal(r.URL.String(), zap.Error(err))
		}

		layout.Base("Products", components.Products(products.Products)).Render(r.Context(), w)
	})

	mw := middlewares.NewMiddleware(
		mux,
		middlewares.WithSecure(ctxClient.Secure),
		middlewares.WithGRPC(ctxClient.GRPCAddress != ""),
	)

	mw = sessionManager.LoadAndSave(mw)

	http.ListenAndServe(addr, mw)
}
