package paymongo

import (
	"context"
	"io"
	"net/http"

	"cchoice/internal/conf"
	"cchoice/internal/database"
	"cchoice/internal/jobs"
	"cchoice/internal/logs"
	"cchoice/internal/payments"

	"github.com/goccy/go-json"
	"go.uber.org/zap"
)

type WebhookEventHandler func(ctx context.Context, event *WebhookEvent)

type WebhookHandlerConfig struct {
	DBRO           database.Service
	DBRW           database.Service
	EmailJobRunner *jobs.EmailJobRunner

	OnPaymentPaid         WebhookEventHandler
	OnPaymentFailed       WebhookEventHandler
	OnCheckoutSessionPaid WebhookEventHandler
}

func NewWebhookHandler(config WebhookHandlerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const logtag = "[PayMongo Webhook Handler]"
		ctx := r.Context()

		body, err := io.ReadAll(r.Body)
		if err != nil {
			logs.LogCtx(ctx).Error(
				logtag,
				zap.String("error", "failed to read request body"),
				zap.Error(err),
			)
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		signatureHeader := r.Header.Get("Paymongo-Signature")
		if signatureHeader == "" {
			logs.LogCtx(ctx).Warn(
				logtag,
				zap.String("error", "missing Paymongo-Signature header"),
			)
			http.Error(w, "Missing signature header", http.StatusUnauthorized)
			return
		}

		signature, err := ParseWebhookSignature(signatureHeader)
		if err != nil {
			logs.LogCtx(ctx).Warn(
				logtag,
				zap.String("error", "invalid signature format"),
				zap.String("signature_header", signatureHeader),
				zap.Error(err),
			)
			http.Error(w, "Invalid signature format", http.StatusUnauthorized)
			return
		}

		if !VerifyWebhookTimestamp(signature.Timestamp, 300) {
			logs.LogCtx(ctx).Warn(
				logtag,
				zap.String("error", "webhook timestamp too old"),
				zap.Int64("timestamp", signature.Timestamp),
			)
		}

		cfg := conf.Conf()
		isLiveMode := IsLiveMode()
		if err := VerifyWebhookSignature(body, signature, cfg.PayMongo.WebhookSecretKey, isLiveMode); err != nil {
			logs.LogCtx(ctx).Warn(
				logtag,
				zap.String("error", "signature verification failed"),
				zap.Bool("is_live_mode", isLiveMode),
				zap.Error(err),
			)
			http.Error(w, "Invalid signature", http.StatusUnauthorized)
			return
		}

		var event WebhookEvent
		if err := json.Unmarshal(body, &event); err != nil {
			logs.LogCtx(ctx).Error(
				logtag,
				zap.String("error", "failed to parse webhook payload"),
				zap.Error(err),
			)
			http.Error(w, "Invalid payload", http.StatusBadRequest)
			return
		}

		logs.LogCtx(ctx).Info(
			logtag,
			zap.String("event_id", event.Data.ID),
			zap.String("event_type", event.Data.Attributes.Type),
			zap.Bool("livemode", event.Data.Attributes.Livemode),
		)

		switch event.Data.Attributes.Type {
		case WebhookEventPaymentPaid:
			handlePaymentPaid(ctx, &event)
			if config.OnPaymentPaid != nil {
				config.OnPaymentPaid(ctx, &event)
			}
		case WebhookEventPaymentFailed:
			handlePaymentFailed(ctx, &event)
			if config.OnPaymentFailed != nil {
				config.OnPaymentFailed(ctx, &event)
			}
		case WebhookEventCheckoutSessionPaymentPaid:
			handleCheckoutSessionPaid(ctx, &event, &config)
			if config.OnCheckoutSessionPaid != nil {
				config.OnCheckoutSessionPaid(ctx, &event)
			}
		default:
			logs.LogCtx(ctx).Info(
				logtag,
				zap.String("action", "unhandled_event_type"),
				zap.String("event_type", event.Data.Attributes.Type),
			)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"statusCode":200,"body":{"message":"SUCCESS"}}`))
	}
}

func handlePaymentPaid(ctx context.Context, event *WebhookEvent) {
	const logtag = "[PayMongo Webhook - Payment Paid]"

	logs.LogCtx(ctx).Info(
		logtag,
		zap.String("event_id", event.Data.ID),
		zap.Any("payment_data", event.Data.Attributes.Data),
	)

	paymentData := event.Data.Attributes.Data
	if paymentData == nil {
		logs.LogCtx(ctx).Warn(logtag, zap.String("error", "payment data is nil"))
		return
	}

	paymentID, _ := paymentData["id"].(string)
	logs.LogCtx(ctx).Info(
		logtag,
		zap.String("payment_id", paymentID),
	)
}

func handlePaymentFailed(ctx context.Context, event *WebhookEvent) {
	const logtag = "[PayMongo Webhook - Payment Failed]"

	logs.LogCtx(ctx).Info(
		logtag,
		zap.String("event_id", event.Data.ID),
		zap.Any("payment_data", event.Data.Attributes.Data),
	)

	paymentData := event.Data.Attributes.Data
	if paymentData == nil {
		logs.LogCtx(ctx).Warn(logtag, zap.String("error", "payment data is nil"))
		return
	}

	failedCode, _ := paymentData["failed_code"].(string)
	failedMessage, _ := paymentData["failed_message"].(string)

	logs.LogCtx(ctx).Warn(
		logtag,
		zap.String("failed_code", failedCode),
		zap.String("failed_message", failedMessage),
	)
}

func handleCheckoutSessionPaid(ctx context.Context, event *WebhookEvent, config *WebhookHandlerConfig) {
	const logtag = "[PayMongo Webhook - Checkout Session Paid]"

	logs.LogCtx(ctx).Info(
		logtag,
		zap.String("event_id", event.Data.ID),
		zap.Any("checkout_data", event.Data.Attributes.Data),
	)

	checkoutData := event.Data.Attributes.Data
	if checkoutData == nil {
		logs.LogCtx(ctx).Warn(logtag, zap.String("error", "checkout data is nil"))
		return
	}

	checkoutSessionID, ok := checkoutData["id"].(string)
	if !ok {
		logs.LogCtx(ctx).Warn(logtag, zap.String("error", "checkout session ID not found"))
		return
	}

	logs.LogCtx(ctx).Info(
		logtag,
		zap.String("checkout_session_id", checkoutSessionID),
	)

	attributes, ok := checkoutData["attributes"].(map[string]any)
	if !ok {
		logs.LogCtx(ctx).Warn(logtag, zap.String("error", "checkout attributes not found"))
		return
	}

	referenceNumber, ok := attributes["reference_number"].(string)
	if !ok || referenceNumber == "" {
		logs.LogCtx(ctx).Warn(logtag, zap.String("error", "reference_number not found in checkout attributes"))
		return
	}

	logs.LogCtx(ctx).Info(
		logtag,
		zap.String("reference_number", referenceNumber),
	)

	if _, err := payments.OnOrderPaid(ctx, payments.OnOrderPaidParams{
		ReferenceNumber: referenceNumber,
		DBRO:            config.DBRO,
		DBRW:            config.DBRW,
		EmailJobRunner:  config.EmailJobRunner,
	}); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
	}
}
