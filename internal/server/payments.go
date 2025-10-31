package server

import (
	"cchoice/cmd/web/components"
	"cchoice/internal/database/queries"
	"cchoice/internal/enums"
	"cchoice/internal/errs"
	"cchoice/internal/logs"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func AddPaymentHandlers(s *Server, r chi.Router) {
	r.Get("/payments/cancel", s.paymentsCancelHandler)
}

func (s *Server) paymentsCancelHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Payments Cancel Handler]"

	orderRefNumber := r.URL.Query().Get("order_ref")
	if orderRefNumber == "" {
		logs.Log().Error(logtag, zap.Error(errs.ErrInvalidParams))
		http.Error(w, "Order reference number is required", http.StatusBadRequest)
		return
	}

	order, err := s.dbRO.GetQueries().GetOrderByOrderNumber(r.Context(), orderRefNumber)
	if err == nil {
		_, err := s.dbRW.GetQueries().UpdateOrderStatus(r.Context(), queries.UpdateOrderStatusParams{
			ID:     order.ID,
			Status: enums.ORDER_STATUS_CANCELLED.String(),
		})
		if err != nil {
			logs.Log().Error(logtag, zap.Error(err), zap.Int64("order_id", order.ID))
		} else {
			logs.Log().Info(logtag, zap.Int64("order_id", order.ID), zap.String("order_number", order.OrderNumber))
		}
	}

	if err := components.CancelPaymentPage(components.CancelPaymentPageBody(orderRefNumber)).Render(r.Context(), w); err != nil {
		logs.Log().Error(logtag, zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
