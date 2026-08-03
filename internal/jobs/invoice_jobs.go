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

type IInvoiceJobService interface {
	GenerateAndStorePDF(ctx context.Context, staffID string, invoiceID int64) error
	SendInvoiceEmailByID(ctx context.Context, staffID string, invoiceID int64) error
}

const (
	InvoiceQueueName      = "invoices"
	JobGenerateInvoicePDF = "generate_invoice_pdf"
	JobSendInvoiceEmail   = "send_invoice_email"
)

type InvoiceJobPayload struct {
	InvoiceJobID int64  `json:"invoice_job_id"`
	StaffID      string `json:"staff_id,omitempty"`
}

type InvoiceJobRunner struct {
	queue             *goqite.Queue
	runner            *jobs.Runner
	dbRO              database.IService
	dbRW              database.IService
	invoiceJobService IInvoiceJobService
}

func NewInvoiceJobRunner(db *sql.DB, dbRO, dbRW database.IService, invoiceJobService IInvoiceJobService) *InvoiceJobRunner {
	if db == nil {
		panic("db is required")
	}
	if invoiceJobService == nil || reflect.ValueOf(invoiceJobService).IsNil() {
		panic("implementor of IInvoiceJobService is required")
	}

	q := goqite.New(goqite.NewOpts{
		DB:   db,
		Name: InvoiceQueueName,
	})

	runner := jobs.NewRunner(jobs.NewRunnerOpts{
		Limit:        3,
		Log:          slog.Default(),
		PollInterval: 5 * time.Second,
		Queue:        q,
	})

	ijr := &InvoiceJobRunner{
		queue:             q,
		runner:            runner,
		dbRO:              dbRO,
		dbRW:              dbRW,
		invoiceJobService: invoiceJobService,
	}

	runner.Register(JobGenerateInvoicePDF, ijr.handleGenerateInvoicePDF)
	runner.Register(JobSendInvoiceEmail, ijr.handleSendInvoiceEmail)

	return ijr
}

func (ijr *InvoiceJobRunner) Start(ctx context.Context) {
	logs.Log().Info("[InvoiceJobRunner] Starting invoice job runner")
	ijr.runner.Start(ctx)
}

func (ijr *InvoiceJobRunner) QueueGeneratePDF(ctx context.Context, invoiceID int64) error {
	return ijr.queueInvoiceJob(ctx, enums.INVOICE_JOB_GENERATE_PDF, invoiceID, "", JobGenerateInvoicePDF)
}

func (ijr *InvoiceJobRunner) QueueSendEmail(ctx context.Context, staffID string, invoiceID int64) error {
	return ijr.queueInvoiceJob(ctx, enums.INVOICE_JOB_SEND_EMAIL, invoiceID, staffID, JobSendInvoiceEmail)
}

func (ijr *InvoiceJobRunner) queueInvoiceJob(
	ctx context.Context,
	jobType enums.InvoiceJobType,
	invoiceID int64,
	staffID string,
	jobName string,
) error {
	const logtag = "[InvoiceJobRunner queueInvoiceJob]"

	tempPayload := InvoiceJobPayload{InvoiceJobID: 0}
	payloadBytes, err := json.Marshal(tempPayload)
	if err != nil {
		return errors.Join(errs.ErrJobsCreateFailed, err)
	}

	if err := ijr.queue.Send(ctx, goqite.Message{Body: payloadBytes}); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		return errors.Join(errs.ErrJobsCreateFailed, err)
	}

	msg, err := ijr.queue.Receive(ctx)
	if err != nil || msg == nil {
		err = cmp.Or(err, errs.ErrJobsNilMessage)
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		return errors.Join(errs.ErrJobsCreateFailed, err)
	}

	queueID := string(msg.ID)

	if err := ijr.queue.Delete(ctx, msg.ID); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		return errors.Join(errs.ErrJobsCreateFailed, err)
	}

	insertParams := queries.InsertInvoiceJobParams{
		QueueID:      queueID,
		InvoiceID:    invoiceID,
		JobType:      jobType.String(),
		Status:       enums.INVOICE_JOB_STATUS_PENDING.String(),
		ErrorMessage: "",
	}

	invoiceJob, err := ijr.dbRW.GetQueries().InsertInvoiceJob(ctx, insertParams)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		return errors.Join(errs.ErrJobsCreateFailed, err)
	}

	payload := InvoiceJobPayload{
		InvoiceJobID: invoiceJob.ID,
		StaffID:      staffID,
	}
	payloadBytes, err = json.Marshal(payload)
	if err != nil {
		return errors.Join(errs.ErrJobsCreateFailed, err)
	}

	if _, err := jobs.Create(ctx, ijr.queue, jobName, goqite.Message{Body: payloadBytes}); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		return errors.Join(errs.ErrJobsCreateFailed, err)
	}

	logs.LogCtx(ctx).Info(
		logtag,
		zap.Int64("invoice_job_id", invoiceJob.ID),
		zap.String("queue_id", queueID),
		zap.Int64("invoice_id", invoiceID),
		zap.Stringer("job_type", jobType),
	)

	return nil
}

func (ijr *InvoiceJobRunner) handleGenerateInvoicePDF(ctx context.Context, m []byte) error {
	const logtag = "[InvoiceJobRunner handleGenerateInvoicePDF]"

	var payload InvoiceJobPayload
	if err := json.Unmarshal(m, &payload); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		return err
	}

	invoiceJobRow, err := ijr.dbRO.GetQueries().GetInvoiceJobByID(ctx, payload.InvoiceJobID)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Int64("invoice_job_id", payload.InvoiceJobID), zap.Error(err))
		return errors.Join(errs.ErrInvoice, err)
	}
	invoiceJob := invoiceJobRow.TblInvoiceJob

	logs.LogCtx(ctx).Info(
		logtag,
		zap.Int64("invoice_job_id", invoiceJob.ID),
		zap.Int64("invoice_id", invoiceJob.InvoiceID),
		zap.String("job_type", invoiceJob.JobType),
	)

	if err := ijr.updateJobStatus(ctx, invoiceJob.ID, enums.INVOICE_JOB_STATUS_PROCESSING, ""); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		return errors.Join(errs.ErrInvoicePDFFailed, err)
	}

	if err := ijr.invoiceJobService.GenerateAndStorePDF(ctx, "", invoiceJob.InvoiceID); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		err2 := ijr.updateJobStatus(ctx, invoiceJob.ID, enums.INVOICE_JOB_STATUS_FAILED, err.Error())
		return errors.Join(errs.ErrInvoicePDFFailed, err, err2)
	}

	if err := ijr.updateJobStatus(ctx, invoiceJob.ID, enums.INVOICE_JOB_STATUS_COMPLETED, ""); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		return errors.Join(errs.ErrInvoicePDFFailed, err)
	}

	logs.LogCtx(ctx).Info(
		logtag,
		zap.String("result", "success"),
		zap.Int64("invoice_id", invoiceJob.InvoiceID),
	)

	return nil
}

func (ijr *InvoiceJobRunner) handleSendInvoiceEmail(ctx context.Context, m []byte) error {
	const logtag = "[InvoiceJobRunner handleSendInvoiceEmail]"

	var payload InvoiceJobPayload
	if err := json.Unmarshal(m, &payload); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		return err
	}

	invoiceJobRow, err := ijr.dbRO.GetQueries().GetInvoiceJobByID(ctx, payload.InvoiceJobID)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Int64("invoice_job_id", payload.InvoiceJobID), zap.Error(err))
		return errors.Join(errs.ErrInvoice, err)
	}
	invoiceJob := invoiceJobRow.TblInvoiceJob

	staffID := payload.StaffID
	if staffID == "" {
		err := errors.New("staff ID is required for send invoice email job")
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		err2 := ijr.updateJobStatus(ctx, invoiceJob.ID, enums.INVOICE_JOB_STATUS_FAILED, err.Error())
		return errors.Join(errs.ErrInvoiceEmailFailed, err, err2)
	}

	logs.LogCtx(ctx).Info(
		logtag,
		zap.Int64("invoice_job_id", invoiceJob.ID),
		zap.Int64("invoice_id", invoiceJob.InvoiceID),
		zap.String("job_type", invoiceJob.JobType),
		zap.String("staff_id", staffID),
	)

	if err := ijr.updateJobStatus(ctx, invoiceJob.ID, enums.INVOICE_JOB_STATUS_PROCESSING, ""); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		return errors.Join(errs.ErrInvoiceEmailFailed, err)
	}

	invoiceRow, err := ijr.dbRO.GetQueries().GetInvoiceByID(ctx, invoiceJob.InvoiceID)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Int64("invoice_id", invoiceJob.InvoiceID), zap.Error(err))
		err2 := ijr.updateJobStatus(ctx, invoiceJob.ID, enums.INVOICE_JOB_STATUS_FAILED, err.Error())
		return errors.Join(errs.ErrInvoiceNotFound, err, err2)
	}

	if invoiceRow.TblInvoice.PdfPath == "" {
		if err := ijr.invoiceJobService.GenerateAndStorePDF(ctx, "", invoiceJob.InvoiceID); err != nil {
			logs.LogCtx(ctx).Error(logtag, zap.Error(err))
			err2 := ijr.updateJobStatus(ctx, invoiceJob.ID, enums.INVOICE_JOB_STATUS_FAILED, err.Error())
			return errors.Join(errs.ErrInvoicePDFFailed, err, err2)
		}
	}

	if err := ijr.invoiceJobService.SendInvoiceEmailByID(ctx, staffID, invoiceJob.InvoiceID); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		err2 := ijr.updateJobStatus(ctx, invoiceJob.ID, enums.INVOICE_JOB_STATUS_FAILED, err.Error())
		return errors.Join(errs.ErrInvoiceEmailFailed, err, err2)
	}

	if err := ijr.updateJobStatus(ctx, invoiceJob.ID, enums.INVOICE_JOB_STATUS_COMPLETED, ""); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		return errors.Join(errs.ErrInvoiceEmailFailed, err)
	}

	logs.LogCtx(ctx).Info(
		logtag,
		zap.String("result", "success"),
		zap.Int64("invoice_id", invoiceJob.InvoiceID),
		zap.String("staff_id", staffID),
	)

	return nil
}

func (ijr *InvoiceJobRunner) updateJobStatus(ctx context.Context, jobID int64, status enums.InvoiceJobStatus, errorMsg string) error {
	const logtag = "[InvoiceJobRunner updateJobStatus]"
	if err := ijr.dbRW.GetQueries().UpdateInvoiceJobStatus(ctx, queries.UpdateInvoiceJobStatusParams{
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
