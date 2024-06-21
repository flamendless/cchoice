package handlers

import (
	"cchoice/internal/logs"
	"net/http"
)

type FnHTTP = func(http.ResponseWriter, *http.Request)
type FnHandler = func(http.ResponseWriter, *http.Request) *HandlerRes


func Validate(fn FnHandler) FnHTTP {
	return func(w http.ResponseWriter, r *http.Request) {
		res := fn(w, r)
		if res.Error != nil {
			logs.LogHTTPHandler(r, res.Error)
			http.Error(w, res.Error.Error(), res.StatusCode)
			return
		}

		res.Component.Render(r.Context(), w)
	}
}
