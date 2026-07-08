package payments

import (
	"context"
	"database/sql"

	"cchoice/internal/conf"
	"cchoice/internal/constants"
	"cchoice/internal/database"
	"cchoice/internal/database/queries"
	"cchoice/internal/enums"
	"cchoice/internal/jobs"
	"cchoice/internal/logs"
	"cchoice/internal/metrics"
	"cchoice/internal/orderhistory"
	"cchoice/internal/utils"

	"go.uber.org/zap"
)

type OnOrderPaidParams struct {
	DBRO            database.IService
	DBRW            database.IService
	EmailJobRunner  *jobs.EmailJobRunner
	ReferenceNumber string
	CPointAwarder   ICPointAwarder
}

type OnOrderPaidResult struct {
	OrderNumber   string
	CustomerEmail string
	OrderID       int64
	EarnedCPoints int64
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
			EarnedCPoints: order.EarnedCpoints,
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

	earnedCPoints := int64(0)
	if params.CPointAwarder != nil && order.CustomerID.Valid {
		earnedCPoints = utils.CalculateOrderEarnedCPoints(order.TotalAmount)
		if earnedCPoints > 0 {
			if _, err := params.CPointAwarder.AwardForPaidOrder(ctx, qtx, order); err != nil {
				logs.LogCtx(ctx).Error(
					logtag,
					zap.Int64("order_id", order.ID),
					zap.String("action", "award_cpoints"),
					zap.Error(err),
				)
				return nil, err
			}
		}
	}

	updatedOrder, err := qtx.UpdateOrderOnPaymentSuccess(ctx, queries.UpdateOrderOnPaymentSuccessParams{
		Status:        enums.ORDER_STATUS_CONFIRMED.String(),
		EarnedCpoints: earnedCPoints,
		ID:            order.ID,
	})
	if err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.Int64("order_id", order.ID),
			zap.Error(err),
		)
		return nil, err
	}

	if err := orderhistory.RecordWithQueries(
		ctx,
		qtx,
		order.ID,
		sql.NullInt64{},
		sql.NullString{String: order.Status, Valid: true},
		enums.ORDER_STATUS_CONFIRMED.String(),
		sql.NullString{},
	); err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.Int64("order_id", order.ID),
			zap.String("action", "record_status_history"),
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

	metrics.Orders.Paid()

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
		EarnedCPoints: updatedOrder.EarnedCpoints,
	}, nil
}
