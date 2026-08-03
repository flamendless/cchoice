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

type ICollectionReceiptJobService interface {
	GenerateAndStorePDF(ctx context.Context, staffID string, receiptID int64) error
	SendCollectionReceiptEmailByID(ctx context.Context, staffID string, receiptID int64) error
}

const (
	CollectionReceiptQueueName      = "collection_receipts"
	JobGenerateCollectionReceiptPDF = "generate_collection_receipt_pdf"
	JobSendCollectionReceiptEmail   = "send_collection_receipt_email"
)

type CollectionReceiptJobPayload struct {
	CollectionReceiptJobID int64  `json:"collection_receipt_job_id"`
	StaffID                string `json:"staff_id,omitempty"`
}

type CollectionReceiptJobRunner struct {
	queue                      *goqite.Queue
	runner                     *jobs.Runner
	dbRO                       database.IService
	dbRW                       database.IService
	collectionReceiptJobService ICollectionReceiptJobService
}

func NewCollectionReceiptJobRunner(db *sql.DB, dbRO, dbRW database.IService, collectionReceiptJobService ICollectionReceiptJobService) *CollectionReceiptJobRunner {
	if db == nil {
		panic("db is required")
	}
	if collectionReceiptJobService == nil || reflect.ValueOf(collectionReceiptJobService).IsNil() {
		panic("implementor of ICollectionReceiptJobService is required")
	}

	q := goqite.New(goqite.NewOpts{
		DB:   db,
		Name: CollectionReceiptQueueName,
	})

	runner := jobs.NewRunner(jobs.NewRunnerOpts{
		Limit:        3,
		Log:          slog.Default(),
		PollInterval: 5 * time.Second,
		Queue:        q,
	})

	crr := &CollectionReceiptJobRunner{
		queue:                      q,
		runner:                     runner,
		dbRO:                       dbRO,
		dbRW:                       dbRW,
		collectionReceiptJobService: collectionReceiptJobService,
	}

	runner.Register(JobGenerateCollectionReceiptPDF, crr.handleGenerateCollectionReceiptPDF)
	runner.Register(JobSendCollectionReceiptEmail, crr.handleSendCollectionReceiptEmail)

	return crr
}

func (crr *CollectionReceiptJobRunner) Start(ctx context.Context) {
	logs.Log().Info("[CollectionReceiptJobRunner] Starting collection receipt job runner")
	crr.runner.Start(ctx)
}

func (crr *CollectionReceiptJobRunner) QueueGeneratePDF(ctx context.Context, receiptID int64) error {
	return crr.queueCollectionReceiptJob(ctx, enums.COLLECTION_RECEIPT_JOB_GENERATE_PDF, receiptID, "", JobGenerateCollectionReceiptPDF)
}

func (crr *CollectionReceiptJobRunner) QueueSendEmail(ctx context.Context, staffID string, receiptID int64) error {
	return crr.queueCollectionReceiptJob(ctx, enums.COLLECTION_RECEIPT_JOB_SEND_EMAIL, receiptID, staffID, JobSendCollectionReceiptEmail)
}

func (crr *CollectionReceiptJobRunner) queueCollectionReceiptJob(
	ctx context.Context,
	jobType enums.CollectionReceiptJobType,
	receiptID int64,
	staffID string,
	jobName string,
) error {
	const logtag = "[CollectionReceiptJobRunner queueCollectionReceiptJob]"

	tempPayload := CollectionReceiptJobPayload{CollectionReceiptJobID: 0}
	payloadBytes, err := json.Marshal(tempPayload)
	if err != nil {
		return errors.Join(errs.ErrJobsCreateFailed, err)
	}

	if err := crr.queue.Send(ctx, goqite.Message{Body: payloadBytes}); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		return errors.Join(errs.ErrJobsCreateFailed, err)
	}

	msg, err := crr.queue.Receive(ctx)
	if err != nil || msg == nil {
		err = cmp.Or(err, errs.ErrJobsNilMessage)
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		return errors.Join(errs.ErrJobsCreateFailed, err)
	}

	queueID := string(msg.ID)

	if err := crr.queue.Delete(ctx, msg.ID); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		return errors.Join(errs.ErrJobsCreateFailed, err)
	}

	insertParams := queries.InsertCollectionReceiptJobParams{
		QueueID:             queueID,
		CollectionReceiptID: receiptID,
		JobType:             jobType.String(),
		Status:              enums.INVOICE_JOB_STATUS_PENDING.String(),
		ErrorMessage:        "",
	}

	receiptJob, err := crr.dbRW.GetQueries().InsertCollectionReceiptJob(ctx, insertParams)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		return errors.Join(errs.ErrJobsCreateFailed, err)
	}

	payload := CollectionReceiptJobPayload{
		CollectionReceiptJobID: receiptJob.ID,
		StaffID:                staffID,
	}
	payloadBytes, err = json.Marshal(payload)
	if err != nil {
		return errors.Join(errs.ErrJobsCreateFailed, err)
	}

	if _, err := jobs.Create(ctx, crr.queue, jobName, goqite.Message{Body: payloadBytes}); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		return errors.Join(errs.ErrJobsCreateFailed, err)
	}

	logs.LogCtx(ctx).Info(
		logtag,
		zap.Int64("collection_receipt_job_id", receiptJob.ID),
		zap.String("queue_id", queueID),
		zap.Int64("collection_receipt_id", receiptID),
		zap.Stringer("job_type", jobType),
	)

	return nil
}

func (crr *CollectionReceiptJobRunner) handleGenerateCollectionReceiptPDF(ctx context.Context, m []byte) error {
	const logtag = "[CollectionReceiptJobRunner handleGenerateCollectionReceiptPDF]"

	var payload CollectionReceiptJobPayload
	if err := json.Unmarshal(m, &payload); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		return err
	}

	receiptJobRow, err := crr.dbRO.GetQueries().GetCollectionReceiptJobByID(ctx, payload.CollectionReceiptJobID)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Int64("collection_receipt_job_id", payload.CollectionReceiptJobID), zap.Error(err))
		return errors.Join(errs.ErrCollectionReceipt, err)
	}
	receiptJob := receiptJobRow.TblCollectionReceiptJob

	logs.LogCtx(ctx).Info(
		logtag,
		zap.Int64("collection_receipt_job_id", receiptJob.ID),
		zap.Int64("collection_receipt_id", receiptJob.CollectionReceiptID),
		zap.String("job_type", receiptJob.JobType),
	)

	if err := crr.updateJobStatus(ctx, receiptJob.ID, enums.INVOICE_JOB_STATUS_PROCESSING, ""); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		return errors.Join(errs.ErrCollectionReceiptPDFFailed, err)
	}

	if err := crr.collectionReceiptJobService.GenerateAndStorePDF(ctx, "", receiptJob.CollectionReceiptID); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		err2 := crr.updateJobStatus(ctx, receiptJob.ID, enums.INVOICE_JOB_STATUS_FAILED, err.Error())
		return errors.Join(errs.ErrCollectionReceiptPDFFailed, err, err2)
	}

	if err := crr.updateJobStatus(ctx, receiptJob.ID, enums.INVOICE_JOB_STATUS_COMPLETED, ""); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		return errors.Join(errs.ErrCollectionReceiptPDFFailed, err)
	}

	logs.LogCtx(ctx).Info(
		logtag,
		zap.String("result", "success"),
		zap.Int64("collection_receipt_id", receiptJob.CollectionReceiptID),
	)

	return nil
}

func (crr *CollectionReceiptJobRunner) handleSendCollectionReceiptEmail(ctx context.Context, m []byte) error {
	const logtag = "[CollectionReceiptJobRunner handleSendCollectionReceiptEmail]"

	var payload CollectionReceiptJobPayload
	if err := json.Unmarshal(m, &payload); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		return err
	}

	receiptJobRow, err := crr.dbRO.GetQueries().GetCollectionReceiptJobByID(ctx, payload.CollectionReceiptJobID)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Int64("collection_receipt_job_id", payload.CollectionReceiptJobID), zap.Error(err))
		return errors.Join(errs.ErrCollectionReceipt, err)
	}
	receiptJob := receiptJobRow.TblCollectionReceiptJob

	staffID := payload.StaffID
	if staffID == "" {
		err := errors.New("staff ID is required for send collection receipt email job")
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		err2 := crr.updateJobStatus(ctx, receiptJob.ID, enums.INVOICE_JOB_STATUS_FAILED, err.Error())
		return errors.Join(errs.ErrCollectionReceiptEmailFailed, err, err2)
	}

	logs.LogCtx(ctx).Info(
		logtag,
		zap.Int64("collection_receipt_job_id", receiptJob.ID),
		zap.Int64("collection_receipt_id", receiptJob.CollectionReceiptID),
		zap.String("job_type", receiptJob.JobType),
		zap.String("staff_id", staffID),
	)

	if err := crr.updateJobStatus(ctx, receiptJob.ID, enums.INVOICE_JOB_STATUS_PROCESSING, ""); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		return errors.Join(errs.ErrCollectionReceiptEmailFailed, err)
	}

	receiptRow, err := crr.dbRO.GetQueries().GetCollectionReceiptByID(ctx, receiptJob.CollectionReceiptID)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Int64("collection_receipt_id", receiptJob.CollectionReceiptID), zap.Error(err))
		err2 := crr.updateJobStatus(ctx, receiptJob.ID, enums.INVOICE_JOB_STATUS_FAILED, err.Error())
		return errors.Join(errs.ErrCollectionReceiptNotFound, err, err2)
	}

	if receiptRow.TblCollectionReceipt.PdfPath == "" {
		if err := crr.collectionReceiptJobService.GenerateAndStorePDF(ctx, "", receiptJob.CollectionReceiptID); err != nil {
			logs.LogCtx(ctx).Error(logtag, zap.Error(err))
			err2 := crr.updateJobStatus(ctx, receiptJob.ID, enums.INVOICE_JOB_STATUS_FAILED, err.Error())
			return errors.Join(errs.ErrCollectionReceiptPDFFailed, err, err2)
		}
	}

	if err := crr.collectionReceiptJobService.SendCollectionReceiptEmailByID(ctx, staffID, receiptJob.CollectionReceiptID); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		err2 := crr.updateJobStatus(ctx, receiptJob.ID, enums.INVOICE_JOB_STATUS_FAILED, err.Error())
		return errors.Join(errs.ErrCollectionReceiptEmailFailed, err, err2)
	}

	if err := crr.updateJobStatus(ctx, receiptJob.ID, enums.INVOICE_JOB_STATUS_COMPLETED, ""); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		return errors.Join(errs.ErrCollectionReceiptEmailFailed, err)
	}

	logs.LogCtx(ctx).Info(
		logtag,
		zap.String("result", "success"),
		zap.Int64("collection_receipt_id", receiptJob.CollectionReceiptID),
		zap.String("staff_id", staffID),
	)

	return nil
}

func (crr *CollectionReceiptJobRunner) updateJobStatus(ctx context.Context, jobID int64, status enums.InvoiceJobStatus, errorMsg string) error {
	const logtag = "[CollectionReceiptJobRunner updateJobStatus]"
	if err := crr.dbRW.GetQueries().UpdateCollectionReceiptJobStatus(ctx, queries.UpdateCollectionReceiptJobStatusParams{
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
