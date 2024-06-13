package client

import (
	"cchoice/internal/ctx"
	"cchoice/internal/logs"
	pb "cchoice/proto"
	"cchoice/client/components"
	"cchoice/client/session"
	"fmt"
	"net/http"
	"time"

	"github.com/a-h/templ"
)

func Serve(ctxSite *ctx.ClientFlags) {
	logs.Log().Info("Starting site server")

	http.Handle("/", templ.Handler(components.Hello("cchoice")))
	http.HandleFunc("/products", func(w http.ResponseWriter, r *http.Request) {
		products := []*pb.Product{}
		components.CProducts(products).Render(r.Context(), w)
	})

	sh := session.NewMiddleware(h, session.WithSecure(ctxSite.Secure))

	server := &http.Server{
		Addr:         fmt.Sprintf("%s%s", ctxSite.Address, ctxSite.Port),
		Handler:      sh,
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 10,
	}
	server.ListenAndServe()
}
