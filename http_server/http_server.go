package http_server

import (
	"cchoice/internal/ctx"
	"cchoice/internal/logs"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

func Serve(ctxAPI ctx.APIFlags) {
	logs.Log().Info("Setting up routers...")
	r := chi.NewRouter()

	logs.Log().Info("Setting up middlewares...")
	r.Use(middleware.Logger)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World!"))
	})

	url := fmt.Sprintf("%s:%d", ctxAPI.Address, ctxAPI.Port)
	logs.Log().Info("Serving...", zap.String("url", url))
	http.ListenAndServe(url, r)
}
