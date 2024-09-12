package handlers

import (
	"cchoice/client/common"
	"cchoice/internal/errs"
	"cchoice/internal/logs"
	"net/http"

	"github.com/a-h/templ"
	"go.uber.org/zap"
)

type FnHTTP = func(http.ResponseWriter, *http.Request)
type FnHandler = func(http.ResponseWriter, *http.Request) *common.HandlerRes

type BaseHandler struct {
	Logger *zap.Logger
}

func NewBaseHandler(logger *zap.Logger) *BaseHandler {
	return &BaseHandler{
		Logger: logger,
	}
}

func (h *BaseHandler) Default(fn FnHandler) FnHTTP {
	return func(w http.ResponseWriter, r *http.Request) {
		res := fn(w, r)
		if res == nil {
			panic("Returned HandlerRes is nil")
		}

		if res.ReplaceURL != "" {
			w.Header().Add("HX-Replace-Url", res.ReplaceURL)
		}

		if res.RedirectTo != "" {
			w.Header().Add("HX-Redirect", res.RedirectTo)
		}

		if res.Error != nil {
			logs.LogHTTPHandler(h.Logger, r, res.Error)

			if res.Error == errs.ERR_NO_AUTH {
				// http.Redirect(w, r, res.RedirectTo, http.StatusTemporaryRedirect)
				w.Header().Add("HX-Redirect", res.RedirectTo)

			} else {
				if res.RedirectTo == "" {
					if res.StatusCode == 0 {
						res.StatusCode = http.StatusInternalServerError
					}

					http.Error(w, res.Error.Error(), res.StatusCode)
				} else {
					w.Header().Add("HX-Redirect", res.RedirectTo)
				}
			}
			return
		}

		if res.Component == nil {
			panic("Returned component in HandlerRes is nil")
		}

		if res.Streaming {
			templ.Handler(res.Component, templ.WithStreaming()).ServeHTTP(w, r)
		} else {
			res.Component.Render(r.Context(), w)
		}
	}
}
