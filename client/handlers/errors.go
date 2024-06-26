package handlers

import (
	"cchoice/client/common"
	"cchoice/internal/logs"
	"net/http"

	"go.uber.org/zap"
)

type FnHTTP = func(http.ResponseWriter, *http.Request)
type FnHandler = func(*http.ResponseWriter, *http.Request) *common.HandlerRes

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
		res := fn(&w, r)
		if res.Error != nil {
			logs.LogHTTPHandler(h.Logger, r, res.Error)
			http.Error(w, res.Error.Error(), res.StatusCode)
			return
		}

		if res.Component == nil {
			panic("Returned component in HandlerRes is nil")
		}

		res.Component.Render(r.Context(), w)
	}
}
