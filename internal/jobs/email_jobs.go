package jobs

import (
	"cchoice/internal/conf"
	"cchoice/internal/constants"
	"cchoice/internal/database"
	"cchoice/internal/database/queries"
	"cchoice/internal/enums"
	"cchoice/internal/errs"
	"cchoice/internal/logs"
	"cchoice/internal/mail"
	"cchoice/internal/utils"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"go.uber.org/zap"
	"maragu.dev/goqite"
	"maragu.dev/goqite/jobs"
)

const (
	EmailQueueName = "emails"
	JobSendEmail   = "send_email"
)

type EmailJobPayload struct {
	EmailJobID int64 `json:"email_job_id"`
}

type EmailJobParams struct {
	Recipient         string
	CC                string
	Subject           string
	TemplateName      enums.EmailTemplateName
	OrderID           *int64
	CheckoutPaymentID *string
}

type EmailJobRunner struct {
	queue       *goqite.Queue
	runner      *jobs.Runner
	dbRO        database.Service
	dbRW        database.Service
	mailService mail.IMailService
}

func NewEmailJobRunner(db *sql.DB, dbRO, dbRW database.Service, mailService mail.IMailService) *EmailJobRunner {
	q := goqite.New(goqite.NewOpts{
		DB:   db,
		Name: EmailQueueName,
	})

	runner := jobs.NewRunner(jobs.NewRunnerOpts{
		Limit:        5,
		Log:          slog.Default(),
		PollInterval: 5 * time.Second,
		Queue:        q,
	})

	ejr := &EmailJobRunner{
		queue:       q,
		runner:      runner,
		dbRO:        dbRO,
		dbRW:        dbRW,
		mailService: mailService,
	}

	runner.Register(JobSendEmail, ejr.handleSendEmail)

	return ejr
}

func (ejr *EmailJobRunner) Start(ctx context.Context) {
	logs.Log().Info("[EmailJobRunner] Starting email job runner")
	ejr.runner.Start(ctx)
}

func (ejr *EmailJobRunner) QueueEmailJob(ctx context.Context, params EmailJobParams) error {
	const logtag = "[EmailJobRunner QueueEmailJob]"

	tempPayload := EmailJobPayload{EmailJobID: 0}
	payloadBytes, err := json.Marshal(tempPayload)
	if err != nil {
		return errors.Join(errs.ErrJobsCreateFailed, err)
	}

	if err := ejr.queue.Send(ctx, goqite.Message{Body: payloadBytes}); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		return errors.Join(errs.ErrJobsCreateFailed, err)
	}

	msg, err := ejr.queue.Receive(ctx)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		return errors.Join(errs.ErrJobsCreateFailed, err)
	}

	queueID := string(msg.ID)

	if err := ejr.queue.Delete(ctx, msg.ID); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		return errors.Join(errs.ErrJobsCreateFailed, err)
	}

	if conf.Conf().IsLocal() {
		params.Subject = "[DEV] - " + params.Subject
	}

	insertParams := queries.InsertEmailJobParams{
		QueueID:      queueID,
		Recipient:    params.Recipient,
		Subject:      params.Subject,
		TemplateName: params.TemplateName.DBValue(),
	}

	if params.CC != "" {
		insertParams.Cc = sql.NullString{String: params.CC, Valid: true}
	}
	if params.OrderID != nil {
		insertParams.OrderID = sql.NullInt64{Int64: *params.OrderID, Valid: true}
	}
	if params.CheckoutPaymentID != nil {
		insertParams.CheckoutPaymentID = sql.NullString{String: *params.CheckoutPaymentID, Valid: true}
	}

	emailJob, err := ejr.dbRW.GetQueries().InsertEmailJob(ctx, insertParams)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		return errors.Join(errs.ErrJobsCreateFailed, err)
	}

	payload := EmailJobPayload{EmailJobID: emailJob.ID}
	payloadBytes, err = json.Marshal(payload)
	if err != nil {
		return errors.Join(errs.ErrJobsCreateFailed, err)
	}

	if err := jobs.Create(ctx, ejr.queue, JobSendEmail, payloadBytes); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		return errors.Join(errs.ErrJobsCreateFailed, err)
	}

	logs.LogCtx(ctx).Info(
		logtag,
		zap.Int64("email_job_id", emailJob.ID),
		zap.String("queue_id", queueID),
		zap.String("recipient", params.Recipient),
		zap.String("template", params.TemplateName.String()),
	)

	return nil
}

func (ejr *EmailJobRunner) handleSendEmail(ctx context.Context, m []byte) error {
	const logtag = "[EmailJobRunner handleSendEmail]"

	var payload EmailJobPayload
	if err := json.Unmarshal(m, &payload); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		return err
	}

	emailJob, err := ejr.dbRO.GetQueries().GetEmailJobByID(ctx, payload.EmailJobID)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Int64("email_job_id", payload.EmailJobID), zap.Error(err))
		return errors.Join(errs.ErrJobsEmailNotFound, err)
	}

	var cc []string
	if emailJob.Cc.Valid && emailJob.Cc.String != "" {
		cc = strings.Split(emailJob.Cc.String, ",")
	}

	recipient := emailJob.Recipient
	templateName := enums.ParseEmailTemplateNameFromDB(emailJob.TemplateName)

	logs.LogCtx(ctx).Info(
		logtag,
		zap.Int64("email_job_id", emailJob.ID),
		zap.String("recipient", recipient),
		zap.Strings("cc", cc),
		zap.String("subject", emailJob.Subject),
		zap.String("template", templateName.String()),
	)

	switch templateName {
	case enums.EMAIL_TEMPLATE_ORDER_CONFIRMATION:
		return ejr.sendOrderConfirmationEmail(ctx, emailJob, recipient, cc, emailJob.Subject)
	case enums.EMAIL_TEMPLATE_PAYMENT_CONFIRMATION:
		return ejr.sendPaymentConfirmationEmail(ctx, emailJob, recipient, cc, emailJob.Subject)
	default:
		err := fmt.Errorf("unknown template: %s", emailJob.TemplateName)
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		return err
	}
}

func (ejr *EmailJobRunner) sendOrderConfirmationEmail(ctx context.Context, emailJob queries.TblEmailJob, recipient string, cc []string, subject string) error {
	const logtag = "[EmailJobRunner sendOrderConfirmationEmail]"

	if !emailJob.OrderID.Valid {
		return errs.ErrJobsOrderNotFound
	}

	order, err := ejr.dbRO.GetQueries().GetOrderByID(ctx, emailJob.OrderID.Int64)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Int64("order_id", emailJob.OrderID.Int64), zap.Error(err))
		return errors.Join(errs.ErrJobsOrderNotFound, err)
	}

	orderLines, err := ejr.dbRO.GetQueries().GetOrderLinesByOrderID(ctx, order.ID)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Int64("order_id", order.ID), zap.Error(err))
		return err
	}

	lineItems := make([]map[string]any, 0, len(orderLines))
	for _, line := range orderLines {
		lineItems = append(lineItems, map[string]any{
			"Name":     line.Name,
			"Quantity": line.Quantity,
			"Price":    utils.NewMoney(line.TotalPrice, line.Currency).Display(),
		})
	}

	shippingAddress := buildAddress(
		order.ShippingAddressLine1,
		order.ShippingAddressLine2,
		order.ShippingCity,
		order.ShippingState,
		order.ShippingPostalCode,
	)

	templateData := mail.TemplateData{
		"LogoURL":          constants.PathEmailLogoCDN,
		"OrderNumber":      order.OrderNumber,
		"PaymentReference": order.CheckoutPaymentID,
		"LineItems":        lineItems,
		"Subtotal":         utils.NewMoney(order.SubtotalAmount, order.Currency).Display(),
		"ShippingFee":      utils.NewMoney(order.ShippingAmount, order.Currency).Display(),
		"Total":            utils.NewMoney(order.TotalAmount, order.Currency).Display(),
		"ShippingAddress":  shippingAddress,
		"DeliveryETA":      order.ShippingEta.String,
	}

	if err := ejr.mailService.SendTemplateEmail(recipient, cc, subject, enums.EMAIL_TEMPLATE_ORDER_CONFIRMATION.FileName(), templateData); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		return errors.Join(errs.ErrJobsSendEmail, err)
	}

	logs.LogCtx(ctx).Info(
		logtag,
		zap.String("result", "success"),
		zap.String("order_number", order.OrderNumber),
		zap.String("recipient", recipient),
		zap.Strings("cc", cc),
	)

	return nil
}

func (ejr *EmailJobRunner) sendPaymentConfirmationEmail(ctx context.Context, emailJob queries.TblEmailJob, recipient string, cc []string, subject string) error {
	const logtag = "[EmailJobRunner sendPaymentConfirmationEmail]"

	if !emailJob.CheckoutPaymentID.Valid {
		return errs.ErrJobsPaymentNotFound
	}

	checkoutPayment, err := ejr.dbRO.GetQueries().GetCheckoutPaymentByID(ctx, emailJob.CheckoutPaymentID.String)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("checkout_payment_id", emailJob.CheckoutPaymentID.String), zap.Error(err))
		return errors.Join(errs.ErrJobsPaymentNotFound, err)
	}

	templateData := mail.TemplateData{
		"PaymentReference": checkoutPayment.ReferenceNumber,
		"Amount":           utils.NewMoney(checkoutPayment.TotalAmount, "PHP").Display(),
		"PaymentMethod":    checkoutPayment.PaymentMethodType,
		"PaidAt":           checkoutPayment.PaidAt.Format("January 2, 2006 3:04 PM"),
	}

	if err := ejr.mailService.SendTemplateEmail(recipient, cc, subject, enums.EMAIL_TEMPLATE_PAYMENT_CONFIRMATION.FileName(), templateData); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		return errors.Join(errs.ErrJobsSendEmail, err)
	}

	logs.LogCtx(ctx).Info(
		logtag,
		zap.String("result", "success"),
		zap.String("payment_reference", checkoutPayment.ReferenceNumber),
		zap.String("recipient", recipient),
		zap.Strings("cc", cc),
	)

	return nil
}

func buildAddress(line1, line2, city, state, postalCode string) string {
	parts := []string{}
	if line1 != "" {
		parts = append(parts, line1)
	}
	if line2 != "" {
		parts = append(parts, line2)
	}
	if city != "" {
		parts = append(parts, city)
	}
	if state != "" {
		parts = append(parts, state)
	}
	if postalCode != "" {
		parts = append(parts, postalCode)
	}
	return strings.Join(parts, ", ")
}
