package site

import (
	"cchoice/internal/ctx"
	"cchoice/internal/logs"
	pb "cchoice/proto"
	"net/http"

	"github.com/a-h/templ"
)

func Serve(ctxSite *ctx.SiteFlags) {
	logs.Log().Info("Starting site server")

	http.Handle("/", templ.Handler(hello("cchoice")))

	http.HandleFunc("/products", func(w http.ResponseWriter, r *http.Request) {
		products := []*pb.Product{}
		cProducts(products).Render(r.Context(), w)
	})

	http.ListenAndServe(ctxSite.Port, nil)
}
