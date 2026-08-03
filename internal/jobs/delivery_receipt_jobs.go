package jobs

import (
	"cmp"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"reflect"
	"time"

	"cchoice/internal/database"
	"cchoice/internal/database/queries"
	"cchoice/internal/enums"
	"cchoice/internal/errs"
	"cchoice/internal/logs"

	"go.uber.org/zap"
	"maragu.dev/goqite"
	"maragu.dev/goqite/jobs"
)

type IDeliveryReceiptJobService interface {
	GenerateAndStorePDF(ctx context.Context, receiptID int64) error
	SendDeliveryReceiptEmailByID(ctx context.Context, staffID string, receiptID int64) error
}

const (
	DeliveryReceiptQueueName           = "delivery_receipts"
	JobGenerateDeliveryReceiptPDF      = "generate_delivery_receipt_pdf"
	JobSendDeliveryReceiptEmail        = "send_delivery_receipt_email"
)

type DeliveryReceiptJobPayload struct {
	DeliveryReceiptJobID int64  `json:"delivery_receipt_job_id"`
	StaffID              string `json:"staff_id,omitempty"`
}

type DeliveryReceiptJobRunner struct {
	queue                    *goqite.Queue
	runner                   *jobs.Runner
	dbRO                     database.IService
	dbRW                     database.IService
	deliveryReceiptJobService IDeliveryReceiptJobService
}

func NewDeliveryReceiptJobRunner(db *sql.DB, dbRO, dbRW database.IService, deliveryReceiptJobService IDeliveryReceiptJobService) *DeliveryReceiptJobRunner {
	if db == nil {
		panic("db is required")
	}
	if deliveryReceiptJobService == nil || reflect.ValueOf(deliveryReceiptJobService).IsNil() {
		panic("implementor of IDeliveryReceiptJobService is required")
	}

	q := goqite.New(goqite.NewOpts{
		DB:   db,
		Name: DeliveryReceiptQueueName,
	})

	runner := jobs.NewRunner(jobs.NewRunnerOpts{
		Limit:        3,
		Log:          slog.Default(),
		PollInterval: 5 * time.Second,
		Queue:        q,
	})

	drjr := &DeliveryReceiptJobRunner{
		queue:                    q,
		runner:                   runner,
		dbRO:                     dbRO,
		dbRW:                     dbRW,
		deliveryReceiptJobService: deliveryReceiptJobService,
	}

	runner.Register(JobGenerateDeliveryReceiptPDF, drjr.handleGenerateDeliveryReceiptPDF)
	runner.Register(JobSendDeliveryReceiptEmail, drjr.handleSendDeliveryReceiptEmail)

	return drjr
}

func (drjr *DeliveryReceiptJobRunner) Start(ctx context.Context) {
	logs.Log().Info("[DeliveryReceiptJobRunner] Starting delivery receipt job runner")
	drjr.runner.Start(ctx)
}

func (drjr *DeliveryReceiptJobRunner) QueueGeneratePDF(ctx context.Context, receiptID int64) error {
	return drjr.queueDeliveryReceiptJob(ctx, enums.DELIVERY_RECEIPT_JOB_GENERATE_PDF, receiptID, "", JobGenerateDeliveryReceiptPDF)
}

func (drjr *DeliveryReceiptJobRunner) QueueSendEmail(ctx context.Context, staffID string, receiptID int64) error {
	return drjr.queueDeliveryReceiptJob(ctx, enums.DELIVERY_RECEIPT_JOB_SEND_EMAIL, receiptID, staffID, JobSendDeliveryReceiptEmail)
}

func (drjr *DeliveryReceiptJobRunner) queueDeliveryReceiptJob(
	ctx context.Context,
	jobType enums.DeliveryReceiptJobType,
	receiptID int64,
	staffID string,
	jobName string,
) error {
	const logtag = "[DeliveryReceiptJobRunner queueDeliveryReceiptJob]"

	tempPayload := DeliveryReceiptJobPayload{DeliveryReceiptJobID: 0}
	payloadBytes, err := json.Marshal(tempPayload)
	if err != nil {
		return errors.Join(errs.ErrJobsCreateFailed, err)
	}

	if err := drjr.queue.Send(ctx, goqite.Message{Body: payloadBytes}); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		return errors.Join(errs.ErrJobsCreateFailed, err)
	}

	msg, err := drjr.queue.Receive(ctx)
	if err != nil || msg == nil {
		err = cmp.Or(err, errs.ErrJobsNilMessage)
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		return errors.Join(errs.ErrJobsCreateFailed, err)
	}

	queueID := string(msg.ID)

	if err := drjr.queue.Delete(ctx, msg.ID); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		return errors.Join(errs.ErrJobsCreateFailed, err)
	}

	insertParams := queries.InsertDeliveryReceiptJobParams{
		QueueID:           queueID,
		DeliveryReceiptID: receiptID,
		JobType:           jobType.String(),
		Status:            enums.INVOICE_JOB_STATUS_PENDING.String(),
		ErrorMessage:      "",
	}

	deliveryReceiptJob, err := drjr.dbRW.GetQueries().InsertDeliveryReceiptJob(ctx, insertParams)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		return errors.Join(errs.ErrJobsCreateFailed, err)
	}

	payload := DeliveryReceiptJobPayload{
		DeliveryReceiptJobID: deliveryReceiptJob.ID,
		StaffID:              staffID,
	}
	payloadBytes, err = json.Marshal(payload)
	if err != nil {
		return errors.Join(errs.ErrJobsCreateFailed, err)
	}

	if _, err := jobs.Create(ctx, drjr.queue, jobName, goqite.Message{Body: payloadBytes}); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		return errors.Join(errs.ErrJobsCreateFailed, err)
	}

	logs.LogCtx(ctx).Info(
		logtag,
		zap.Int64("delivery_receipt_job_id", deliveryReceiptJob.ID),
		zap.String("queue_id", queueID),
		zap.Int64("delivery_receipt_id", receiptID),
		zap.Stringer("job_type", jobType),
	)

	return nil
}

func (drjr *DeliveryReceiptJobRunner) handleGenerateDeliveryReceiptPDF(ctx context.Context, m []byte) error {
	const logtag = "[DeliveryReceiptJobRunner handleGenerateDeliveryReceiptPDF]"

	var payload DeliveryReceiptJobPayload
	if err := json.Unmarshal(m, &payload); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		return err
	}

	jobRow, err := drjr.dbRO.GetQueries().GetDeliveryReceiptJobByID(ctx, payload.DeliveryReceiptJobID)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Int64("delivery_receipt_job_id", payload.DeliveryReceiptJobID), zap.Error(err))
		return errors.Join(errs.ErrDeliveryReceipt, err)
	}
	deliveryReceiptJob := jobRow.TblDeliveryReceiptJob

	logs.LogCtx(ctx).Info(
		logtag,
		zap.Int64("delivery_receipt_job_id", deliveryReceiptJob.ID),
		zap.Int64("delivery_receipt_id", deliveryReceiptJob.DeliveryReceiptID),
		zap.String("job_type", deliveryReceiptJob.JobType),
	)

	if err := drjr.updateJobStatus(ctx, deliveryReceiptJob.ID, enums.INVOICE_JOB_STATUS_PROCESSING, ""); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		return errors.Join(errs.ErrDeliveryReceiptPDFFailed, err)
	}

	if err := drjr.deliveryReceiptJobService.GenerateAndStorePDF(ctx, deliveryReceiptJob.DeliveryReceiptID); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		err2 := drjr.updateJobStatus(ctx, deliveryReceiptJob.ID, enums.INVOICE_JOB_STATUS_FAILED, err.Error())
		return errors.Join(errs.ErrDeliveryReceiptPDFFailed, err, err2)
	}

	if err := drjr.updateJobStatus(ctx, deliveryReceiptJob.ID, enums.INVOICE_JOB_STATUS_COMPLETED, ""); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		return errors.Join(errs.ErrDeliveryReceiptPDFFailed, err)
	}

	logs.LogCtx(ctx).Info(
		logtag,
		zap.String("result", "success"),
		zap.Int64("delivery_receipt_id", deliveryReceiptJob.DeliveryReceiptID),
	)

	return nil
}

func (drjr *DeliveryReceiptJobRunner) handleSendDeliveryReceiptEmail(ctx context.Context, m []byte) error {
	const logtag = "[DeliveryReceiptJobRunner handleSendDeliveryReceiptEmail]"

	var payload DeliveryReceiptJobPayload
	if err := json.Unmarshal(m, &payload); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		return err
	}

	jobRow, err := drjr.dbRO.GetQueries().GetDeliveryReceiptJobByID(ctx, payload.DeliveryReceiptJobID)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Int64("delivery_receipt_job_id", payload.DeliveryReceiptJobID), zap.Error(err))
		return errors.Join(errs.ErrDeliveryReceipt, err)
	}
	deliveryReceiptJob := jobRow.TblDeliveryReceiptJob

	staffID := payload.StaffID
	if staffID == "" {
		err := errors.New("staff ID is required for send delivery receipt email job")
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		err2 := drjr.updateJobStatus(ctx, deliveryReceiptJob.ID, enums.INVOICE_JOB_STATUS_FAILED, err.Error())
		return errors.Join(errs.ErrDeliveryReceiptEmailFailed, err, err2)
	}

	logs.LogCtx(ctx).Info(
		logtag,
		zap.Int64("delivery_receipt_job_id", deliveryReceiptJob.ID),
		zap.Int64("delivery_receipt_id", deliveryReceiptJob.DeliveryReceiptID),
		zap.String("job_type", deliveryReceiptJob.JobType),
		zap.String("staff_id", staffID),
	)

	if err := drjr.updateJobStatus(ctx, deliveryReceiptJob.ID, enums.INVOICE_JOB_STATUS_PROCESSING, ""); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		return errors.Join(errs.ErrDeliveryReceiptEmailFailed, err)
	}

	receiptRow, err := drjr.dbRO.GetQueries().GetDeliveryReceiptByID(ctx, deliveryReceiptJob.DeliveryReceiptID)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Int64("delivery_receipt_id", deliveryReceiptJob.DeliveryReceiptID), zap.Error(err))
		err2 := drjr.updateJobStatus(ctx, deliveryReceiptJob.ID, enums.INVOICE_JOB_STATUS_FAILED, err.Error())
		return errors.Join(errs.ErrDeliveryReceiptNotFound, err, err2)
	}

	if receiptRow.TblDeliveryReceipt.PdfPath == "" {
		if err := drjr.deliveryReceiptJobService.GenerateAndStorePDF(ctx, deliveryReceiptJob.DeliveryReceiptID); err != nil {
			logs.LogCtx(ctx).Error(logtag, zap.Error(err))
			err2 := drjr.updateJobStatus(ctx, deliveryReceiptJob.ID, enums.INVOICE_JOB_STATUS_FAILED, err.Error())
			return errors.Join(errs.ErrDeliveryReceiptPDFFailed, err, err2)
		}
	}

	if err := drjr.deliveryReceiptJobService.SendDeliveryReceiptEmailByID(ctx, staffID, deliveryReceiptJob.DeliveryReceiptID); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		err2 := drjr.updateJobStatus(ctx, deliveryReceiptJob.ID, enums.INVOICE_JOB_STATUS_FAILED, err.Error())
		return errors.Join(errs.ErrDeliveryReceiptEmailFailed, err, err2)
	}

	if err := drjr.updateJobStatus(ctx, deliveryReceiptJob.ID, enums.INVOICE_JOB_STATUS_COMPLETED, ""); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		return errors.Join(errs.ErrDeliveryReceiptEmailFailed, err)
	}

	logs.LogCtx(ctx).Info(
		logtag,
		zap.String("result", "success"),
		zap.Int64("delivery_receipt_id", deliveryReceiptJob.DeliveryReceiptID),
		zap.String("staff_id", staffID),
	)

	return nil
}

func (drjr *DeliveryReceiptJobRunner) updateJobStatus(ctx context.Context, jobID int64, status enums.InvoiceJobStatus, errorMsg string) error {
	const logtag = "[DeliveryReceiptJobRunner updateJobStatus]"
	if err := drjr.dbRW.GetQueries().UpdateDeliveryReceiptJobStatus(ctx, queries.UpdateDeliveryReceiptJobStatusParams{
		Status:       status.String(),
		ErrorMessage: errorMsg,
		ID:           jobID,
	}); err != nil {
		logs.Log().Warn(
			logtag,
			zap.Int64("job id", jobID),
			zap.String("status", status.String()),
			zap.String("error message", errorMsg),
			zap.Error(err),
		)
		return err
	}
	return nil
}
