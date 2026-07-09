package server

import (
	"cmp"
	"database/sql"
	"net/http"

	comppayment "cchoice/cmd/web/components/payment"
	"cchoice/internal/conf"
	"cchoice/internal/database/queries"
	"cchoice/internal/enums"
	"cchoice/internal/errs"
	"cchoice/internal/httputil"
	"cchoice/internal/logs"
	"cchoice/internal/metrics"
	"cchoice/internal/orderhistory"
	"cchoice/internal/payments"
	"cchoice/internal/payments/paymongo"
	"cchoice/internal/server/forms"
	"cchoice/internal/utils"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func AddPaymentHandlers(s *Server, r chi.Router) {
	r.Get("/payments/cancel", s.paymentsCancelHandler)
	r.Get("/payments/success", s.paymentsSuccessHandler)
}

func RegisterPaymentWebhooks(s *Server, r chi.Router) {
	if s.paymentGateway == nil {
		return
	}
	if s.paymentGateway.GatewayEnum() == payments.PAYMENT_GATEWAY_PAYMONGO {
		handler := paymongo.NewWebhookHandler(paymongo.WebhookHandlerConfig{
			DBRO:           s.dbRO,
			DBRW:           s.dbRW,
			EmailJobRunner: s.mailJobRunner,
			CPointAwarder:  s.services.cpoint,
		})
		r.Post("/webhooks/paymongo", handler)
		return
	}

	logs.Log().Warn("No payment webhooks registered")
}

func (s *Server) paymentsCancelHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Payments Cancel Handler]"
	ctx := r.Context()

	var q forms.PaymentRedirectQuery
	if err := httputil.BindQuery(r, &q); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(errs.ErrInvalidParams))
		http.Error(w, "Payment reference number is required", http.StatusBadRequest)
		return
	}
	paymentRef := q.PaymentRef

	webhookSecret := conf.Conf().PayMongo.WebhookSecretKey
	tokenValid := webhookSecret != "" && payments.VerifyRedirectToken(webhookSecret, paymentRef, q.Token)

	logs.LogCtx(ctx).Info(logtag, zap.String("payment_ref", paymentRef), zap.Bool("token_valid", tokenValid))

	if tokenValid {
		if checkoutPayment, err := s.dbRO.GetQueries().GetCheckoutPaymentByReferenceNumber(ctx, paymentRef); err == nil {
			if order, err := s.dbRO.GetQueries().GetOrderByCheckoutPaymentID(ctx, checkoutPayment.ID); err == nil {
				if enums.ParseOrderStatusToEnum(order.Status) != enums.ORDER_STATUS_PENDING {
					logs.LogCtx(ctx).Warn(
						logtag,
						zap.Int64("order_id", order.ID),
						zap.String("status", order.Status),
						zap.String("action", "skip_cancel_non_pending"),
					)
				} else {
					previousStatus := order.Status
					if _, err := s.dbRW.GetQueries().UpdateOrderStatus(ctx, queries.UpdateOrderStatusParams{
						ID:     order.ID,
						Status: enums.ORDER_STATUS_CANCELLED.String(),
					}); err != nil {
						logs.LogCtx(ctx).Error(
							logtag,
							zap.Int64("order_id", order.ID),
							zap.Error(err),
						)
					} else if err := orderhistory.Record(
						ctx,
						s.dbRW,
						order.ID,
						sql.NullInt64{},
						sql.NullString{String: previousStatus, Valid: true},
						enums.ORDER_STATUS_CANCELLED.String(),
						sql.NullString{},
					); err != nil {
						logs.LogCtx(ctx).Error(
							logtag,
							zap.Int64("order_id", order.ID),
							zap.String("action", "record_status_history"),
							zap.Error(err),
						)
					} else {
						metrics.Orders.Cancelled()
						if s.mailJobRunner != nil {
							if updatedOrder, err := s.dbRO.GetQueries().GetOrderByID(ctx, order.ID); err != nil {
								logs.LogCtx(ctx).Error(
									logtag,
									zap.Int64("order_id", order.ID),
									zap.String("action", "load_order_for_status_email"),
									zap.Error(err),
								)
							} else if err := s.mailJobRunner.QueueOrderStatusUpdateEmail(ctx, updatedOrder); err != nil {
								logs.LogCtx(ctx).Error(
									logtag,
									zap.Int64("order_id", order.ID),
									zap.String("action", "queue_status_email"),
									zap.Error(err),
								)
							}
						}
					}
				}
			}
		}
	}

	if err := comppayment.CancelPaymentPage(comppayment.CancelPaymentPageBody(paymentRef)).Render(ctx, w); err != nil {
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

	var q forms.PaymentRedirectQuery
	if err := httputil.BindQuery(r, &q); err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.Error(errs.ErrInvalidParams),
		)
		http.Error(w, "Payment reference number is required", http.StatusBadRequest)
		return
	}
	paymentRef := q.PaymentRef

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

	logs.LogCtx(ctx).Info(
		logtag,
		zap.String("payment_ref", paymentRef),
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
		cancelURL := utils.URL("/payments/cancel?payment_ref=" + paymentRef)
		if q.Token != "" {
			cancelURL = utils.URL("/payments/cancel?payment_ref=" + paymentRef + "&token=" + q.Token)
		}
		http.Redirect(w, r, cancelURL, http.StatusSeeOther)
		return
	}

	result, err := payments.OnOrderPaid(ctx, payments.OnOrderPaidParams{
		ReferenceNumber: paymentRef,
		DBRO:            s.dbRO,
		DBRW:            s.dbRW,
		EmailJobRunner:  s.mailJobRunner,
		CPointAwarder:   s.services.cpoint,
	})
	if err != nil || result == nil {
		err = cmp.Or(err, errs.ErrRespNil)
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		http.Error(w, "Failed to process payment", http.StatusInternalServerError)
		return
	}

	if err := comppayment.SuccessPaymentPage(comppayment.SuccessPaymentPageBody(result.OrderNumber, result.EarnedCPoints)).Render(ctx, w); err != nil {
		logs.Log().Error(logtag, zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
