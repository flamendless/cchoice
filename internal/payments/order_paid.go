package payments

import (
	"context"

	"cchoice/internal/conf"
	"cchoice/internal/constants"
	"cchoice/internal/database"
	"cchoice/internal/database/queries"
	"cchoice/internal/enums"
	"cchoice/internal/jobs"
	"cchoice/internal/logs"

	"go.uber.org/zap"
)

type OnOrderPaidParams struct {
	ReferenceNumber string
	DBRO            database.Service
	DBRW            database.Service
	EmailJobRunner  *jobs.EmailJobRunner
}

type OnOrderPaidResult struct {
	OrderNumber   string
	OrderID       int64
	CustomerEmail string
}

func OnOrderPaid(ctx context.Context, params OnOrderPaidParams) (*OnOrderPaidResult, error) {
	const logtag = "[OnOrderPaid]"

	if params.DBRO == nil || params.DBRW == nil {
		logs.LogCtx(ctx).Warn(logtag, zap.String("error", "database services not configured"))
		return nil, nil
	}

	checkoutPayment, err := params.DBRO.GetQueries().GetCheckoutPaymentByReferenceNumber(ctx, params.ReferenceNumber)
	if err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.String("reference_number", params.ReferenceNumber),
			zap.Error(err),
		)
		return nil, err
	}

	if checkoutPayment.Status == enums.PAYMENT_STATUS_PAID.String() {
		logs.LogCtx(ctx).Info(
			logtag,
			zap.String("action", "already_processed"),
			zap.String("checkout_payment_id", checkoutPayment.ID),
		)
		order, err := params.DBRO.GetQueries().GetOrderByCheckoutPaymentID(ctx, checkoutPayment.ID)
		if err != nil {
			return nil, err
		}
		return &OnOrderPaidResult{
			OrderNumber:   order.OrderNumber,
			OrderID:       order.ID,
			CustomerEmail: order.CustomerEmail,
		}, nil
	}

	order, err := params.DBRO.GetQueries().GetOrderByCheckoutPaymentID(ctx, checkoutPayment.ID)
	if err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.String("checkout_payment_id", checkoutPayment.ID),
			zap.Error(err),
		)
		return nil, err
	}

	tx, err := params.DBRW.GetDB().BeginTx(ctx, nil)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		return nil, err
	}
	defer func() {
		if err := tx.Rollback(); err != nil {
			logs.LogCtx(ctx).Debug(logtag, zap.Error(err))
		}
	}()

	qtx := params.DBRW.GetQueries().WithTx(tx)

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
		return nil, err
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
		return nil, err
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
		return nil, err
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
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		return nil, err
	}

	logs.LogCtx(ctx).Info(
		logtag,
		zap.String("result", "success"),
		zap.String("reference_number", params.ReferenceNumber),
		zap.String("order_number", updatedOrder.OrderNumber),
		zap.Int64("order_id", updatedOrder.ID),
	)

	if params.EmailJobRunner != nil {
		cfg := conf.Conf()
		if err := params.EmailJobRunner.QueueEmailJob(ctx, jobs.EmailJobParams{
			Recipient:         updatedOrder.CustomerEmail,
			CC:                cfg.MailerooConfig.CC,
			Subject:           "Order Confirmation - " + updatedOrder.OrderNumber,
			TemplateName:      enums.EMAIL_TEMPLATE_ORDER_CONFIRMATION,
			OrderID:           &updatedOrder.ID,
			CheckoutPaymentID: &updatedOrder.CheckoutPaymentID,
			MobileNo:          constants.ViberURIPrefix + cfg.Settings.MobileNo,
			EMail:             cfg.Settings.EMail,
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
	}

	return &OnOrderPaidResult{
		OrderNumber:   updatedOrder.OrderNumber,
		OrderID:       updatedOrder.ID,
		CustomerEmail: updatedOrder.CustomerEmail,
	}, nil
}
