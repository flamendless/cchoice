package server

import (
	"cchoice/cmd/web/components"
	"cchoice/internal/conf"
	"cchoice/internal/database/queries"
	"cchoice/internal/enums"
	"cchoice/internal/errs"
	"cchoice/internal/jobs"
	"cchoice/internal/logs"
	"cchoice/internal/payments"
	"cchoice/internal/payments/paymongo"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func AddPaymentHandlers(s *Server, r chi.Router) {
	r.Get("/payments/cancel", s.paymentsCancelHandler)
	r.Get("/payments/success", s.paymentsSuccessHandler)
}

func (s *Server) paymentsCancelHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Payments Cancel Handler]"
	ctx := r.Context()

	paymentRef := r.URL.Query().Get("payment_ref")
	if paymentRef == "" {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.Error(errs.ErrInvalidParams),
		)
		http.Error(w, "Payment reference number is required", http.StatusBadRequest)
		return
	}

	logs.LogCtx(ctx).Info(
		logtag,
		zap.String("payment_ref", paymentRef),
		zap.String("query_params", r.URL.RawQuery),
	)

	if checkoutPayment, err := s.dbRO.GetQueries().GetCheckoutPaymentByReferenceNumber(ctx, paymentRef); err == nil {
		if order, err := s.dbRO.GetQueries().GetOrderByCheckoutPaymentID(ctx, checkoutPayment.ID); err == nil {
			if _, err := s.dbRW.GetQueries().UpdateOrderStatus(ctx, queries.UpdateOrderStatusParams{
				ID:     order.ID,
				Status: enums.ORDER_STATUS_CANCELLED.String(),
			}); err != nil {
				logs.LogCtx(ctx).Error(
					logtag,
					zap.Int64("order_id", order.ID),
					zap.Error(err),
				)
			}
		}
	}

	if err := components.CancelPaymentPage(components.CancelPaymentPageBody(paymentRef)).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) paymentsSuccessHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Payments Success Handler]"
	ctx := r.Context()

	paymentRef := r.URL.Query().Get("payment_ref")
	if paymentRef == "" {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.Error(errs.ErrInvalidParams),
		)
		http.Error(w, "Payment reference number is required", http.StatusBadRequest)
		return
	}

	logs.LogCtx(ctx).Info(
		logtag,
		zap.String("payment_ref", paymentRef),
		zap.String("query_params", r.URL.RawQuery),
	)

	checkoutPayment, err := s.dbRO.GetQueries().GetCheckoutPaymentByReferenceNumber(ctx, paymentRef)
	if err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.String("payment_ref", paymentRef),
			zap.Error(err),
		)
		http.Error(w, "Payment information not found", http.StatusNotFound)
		return
	}

	order, err := s.dbRO.GetQueries().GetOrderByCheckoutPaymentID(ctx, checkoutPayment.ID)
	if err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.String("checkout_payment_id", checkoutPayment.ID),
			zap.Error(err),
		)
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	logs.LogCtx(ctx).Info(
		logtag,
		zap.String("checkout_payment_id", checkoutPayment.ID),
		zap.String("payment_intent_id", checkoutPayment.PaymentIntentID.String),
		zap.String("payment_status", checkoutPayment.Status),
	)

	if !checkoutPayment.PaymentIntentID.Valid || checkoutPayment.PaymentIntentID.String == "" {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.String("payment_ref", paymentRef),
			zap.Error(errs.ErrPaymentResponse),
		)
		http.Error(w, "Payment intent ID not found", http.StatusBadRequest)
		return
	}

	var paymentStatus string
	switch s.paymentGateway.GatewayEnum() {
	case payments.PAYMENT_GATEWAY_PAYMONGO:
		paymongoGateway, ok := s.paymentGateway.(*paymongo.PayMongo)
		if !ok {
			logs.LogCtx(ctx).Error(
				logtag,
				zap.Error(errs.ErrServerUnimplementedGateway),
			)
			http.Error(w, "Payment gateway type mismatch", http.StatusInternalServerError)
			return
		}

		paymentIntentRes, err := paymongoGateway.GetPaymentIntent(checkoutPayment.PaymentIntentID.String)
		if err != nil {
			logs.LogCtx(ctx).Error(
				logtag,
				zap.String("payment_intent_id", checkoutPayment.PaymentIntentID.String),
				zap.Error(err),
			)
			http.Error(w, "Failed to verify payment", http.StatusInternalServerError)
			return
		}

		logs.LogCtx(ctx).Info(
			logtag,
			zap.String("payment_intent_id", paymentIntentRes.Data.ID),
			zap.String("payment_intent_status", paymentIntentRes.Data.Attributes.Status),
			zap.Any("payment_intent_response", paymentIntentRes),
		)

		paymentStatus = paymentIntentRes.Data.Attributes.Status

	default:
		err := errs.ErrServerUnimplementedGateway
		logs.LogCtx(ctx).Error(
			logtag,
			zap.String("gateway", s.paymentGateway.GatewayEnum().String()),
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusNotImplemented)
		return
	}

	if paymentStatus != "succeeded" {
		logs.LogCtx(ctx).Warn(
			logtag,
			zap.String("payment_ref", paymentRef),
			zap.String("gateway", s.paymentGateway.GatewayEnum().String()),
			zap.String("expected_status", "succeeded"),
			zap.String("actual_status", paymentStatus),
		)
		http.Redirect(w, r, "/cchoice/payments/cancel?payment_ref="+paymentRef, http.StatusSeeOther)
		return
	}

	tx, err := s.dbRW.GetDB().BeginTx(ctx, nil)
	if err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.Error(err),
		)
		http.Error(w, "Failed to process payment", http.StatusInternalServerError)
		return
	}
	defer func() {
		if err := tx.Rollback(); err != nil {
			logs.LogCtx(ctx).Error(
				logtag,
				zap.Error(err),
			)
		}
	}()

	qtx := s.dbRW.GetQueries().WithTx(tx)

	updatedCheckoutPayment, err := qtx.UpdateCheckoutPaymentOnSuccess(ctx, queries.UpdateCheckoutPaymentOnSuccessParams{
		Status: enums.PAYMENT_STATUS_PAID.String(),
		ID:     checkoutPayment.ID,
	})
	if err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.String("checkout_payment_id", checkoutPayment.ID),
			zap.Error(err),
		)
		http.Error(w, "Failed to update payment status", http.StatusInternalServerError)
		return
	}
	logs.LogCtx(ctx).Info(
		logtag,
		zap.String("action", "updated_checkout_payment"),
		zap.String("checkout_payment_id", updatedCheckoutPayment.ID),
		zap.String("new_status", updatedCheckoutPayment.Status),
	)

	updatedCheckout, err := qtx.UpdateCheckoutStatus(ctx, queries.UpdateCheckoutStatusParams{
		Status: enums.CHECKOUT_STATUS_COMPLETED.String(),
		ID:     order.CheckoutID,
	})
	if err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.Int64("checkout_id", order.CheckoutID),
			zap.Error(err),
		)
		http.Error(w, "Failed to update checkout status", http.StatusInternalServerError)
		return
	}
	logs.LogCtx(ctx).Info(
		logtag,
		zap.String("action", "updated_checkout"),
		zap.Int64("checkout_id", updatedCheckout.ID),
		zap.String("new_status", updatedCheckout.Status),
	)

	updatedOrder, err := qtx.UpdateOrderOnPaymentSuccess(ctx, queries.UpdateOrderOnPaymentSuccessParams{
		Status: enums.ORDER_STATUS_CONFIRMED.String(),
		ID:     order.ID,
	})
	if err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.Int64("order_id", order.ID),
			zap.Error(err),
		)
		http.Error(w, "Failed to update order status", http.StatusInternalServerError)
		return
	}
	logs.LogCtx(ctx).Info(
		logtag,
		zap.String("action", "updated_order"),
		zap.Int64("order_id", updatedOrder.ID),
		zap.String("order_number", updatedOrder.OrderNumber),
		zap.String("new_status", updatedOrder.Status),
		zap.Time("paid_at", updatedOrder.PaidAt.Time),
	)

	if err := tx.Commit(); err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.Error(err),
		)
		http.Error(w, "Failed to finalize payment", http.StatusInternalServerError)
		return
	}

	logs.LogCtx(ctx).Info(
		logtag,
		zap.String("result", "success"),
		zap.String("payment_ref", paymentRef),
		zap.String("order_number", updatedOrder.OrderNumber),
		zap.Int64("order_id", updatedOrder.ID),
	)

	if err := s.emailJobRunner.QueueEmailJob(ctx, jobs.EmailJobParams{
		Recipient:         updatedOrder.CustomerEmail,
		CC:                conf.Conf().MailerooConfig.CC,
		Subject:           "Order Confirmation - " + updatedOrder.OrderNumber,
		TemplateName:      enums.EMAIL_TEMPLATE_ORDER_CONFIRMATION,
		OrderID:           &updatedOrder.ID,
		CheckoutPaymentID: &updatedOrder.CheckoutPaymentID,
	}); err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.String("action", "queue_email_failed"),
			zap.Int64("order_id", updatedOrder.ID),
			zap.Error(err),
		)
	} else {
		logs.LogCtx(ctx).Info(
			logtag,
			zap.String("action", "email_queued"),
			zap.Int64("order_id", updatedOrder.ID),
			zap.String("recipient", updatedOrder.CustomerEmail),
		)
	}

	if err := components.SuccessPaymentPage(components.SuccessPaymentPageBody(updatedOrder.OrderNumber)).Render(r.Context(), w); err != nil {
		logs.Log().Error(logtag, zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
