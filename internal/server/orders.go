package server

import (
	"cchoice/cmd/web/components"
	"cchoice/internal/conf"
	"cchoice/internal/constants"
	"cchoice/internal/logs"
	"net/http"
	"fmt"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func AddOrdersHandlers(s *Server, r chi.Router) {
	r.Get("/orders/track", s.ordersTrackPageHandler)
}

func (s *Server) ordersTrackPageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Order Track Page Handler]"
	cfg := conf.Conf()
	email := "mailto:" + cfg.Settings.EMail
	mobileNo := constants.ViberURIPrefix + cfg.Settings.MobileNo
	ctx := r.Context()

	orderNo := r.URL.Query().Get("order_no")
	if orderNo == "" {
		if err := components.OrderTrackerPage(components.OrderTrackerPageBody(orderNo, email, mobileNo)).Render(ctx, w); err != nil {
			logs.Log().Error(logtag, zap.String("order_no", orderNo), zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	logs.LogCtx(ctx).Info(logtag, zap.String("order_no", orderNo))

	order, err := s.dbRO.GetQueries().GetOrderByOrderNumber(ctx, orderNo)
	if err != nil {
		if err := components.OrderTrackerPage(components.OrderTrackerPageBodyError(orderNo, email, mobileNo)).Render(ctx, w); err != nil {
			logs.Log().Error(logtag, zap.String("order_no", orderNo), zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}

	fmt.Println(order)

	if err := components.OrderTrackerPage(components.OrderTrackerPageBody(orderNo, email, mobileNo)).Render(ctx, w); err != nil {
		logs.Log().Error(logtag, zap.String("order_no", orderNo), zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
