package client

import (
	"cchoice/client/components"
	"cchoice/client/middlewares"
	"cchoice/internal/ctx"
	"cchoice/internal/logs"
	pb "cchoice/proto"
	"fmt"
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
	"go.uber.org/zap"
)

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

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		components.Hello("cchoice").Render(r.Context(), w)
	})
	// mux.HandleFunc("GET /", getHandler)
	// mux.HandleFunc("PUT /", putHandler)
	mux.HandleFunc("GET /products", func(w http.ResponseWriter, r *http.Request) {
		products := []*pb.Product{}
		components.Products(products).Render(r.Context(), w)
	})

	mw := middlewares.NewMiddleware(mux)
	http.ListenAndServe(addr, sessionManager.LoadAndSave(mw))
}
