package handlers

import (
	"cchoice/client/common"
	"cchoice/internal/errs"
	"cchoice/internal/logs"
	"net/http"

	"go.uber.org/zap"
)

type FnHTTP = func(http.ResponseWriter, *http.Request)
type FnHandler = func(http.ResponseWriter, *http.Request) *common.HandlerRes

type ErrorHandler struct {
	Logger *zap.Logger
}

func NewErrorHandler(logger *zap.Logger) *ErrorHandler {
	return &ErrorHandler{
		Logger: logger,
	}
}

func (h *ErrorHandler) Default(fn FnHandler) FnHTTP {
	return func(w http.ResponseWriter, r *http.Request) {
		res := fn(w, r)

		if res == nil {
			panic("Returned HandlerRes is nil")
		}

		if res.ReplaceURL != "" {
			w.Header().Add("HX-Replace-Url", res.ReplaceURL)
		}

		if res.Error != nil {
			logs.LogHTTPHandler(h.Logger, r, res.Error)

			if res.Error == errs.ERR_NO_AUTH {
				http.Redirect(w, r, res.RedirectTo, http.StatusTemporaryRedirect)

			} else {
				if res.RedirectTo == "" {
					if res.StatusCode == 0 {
						res.StatusCode = http.StatusInternalServerError
					}

					http.Error(w, res.Error.Error(), res.StatusCode)
				} else {
					http.Redirect(w, r, res.RedirectTo, http.StatusTemporaryRedirect)
				}
			}
			return
		}

		if res.Component == nil {
			panic("Returned component in HandlerRes is nil")
		}

		res.Component.Render(r.Context(), w)
	}
}
