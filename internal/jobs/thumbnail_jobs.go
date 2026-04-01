package jobs

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"reflect"
	"time"

	"cchoice/internal/conf"
	"cchoice/internal/database"
	"cchoice/internal/database/queries"
	"cchoice/internal/errs"
	"cchoice/internal/logs"
	"cchoice/internal/types"

	"go.uber.org/zap"
	"maragu.dev/goqite"
	"maragu.dev/goqite/jobs"
)

type IThumbnailService interface {
	ProcessImageVariants(ctx context.Context, sourcePath, brand, filename string) ([]types.ThumbnailVariant, error)
}

const (
	ThumbnailQueueName = "thumbnails"
	JobCreateThumbnail = "create_thumbnail"
)

type ThumbnailJobPayload struct {
	ThumbnailJobID int64 `json:"thumbnail_job_id"`
}

type ThumbnailJobParams struct {
	ProductID  int64
	Brand      string
	SourcePath string
	Filename   string
}

type ThumbnailJobRunner struct {
	queue            *goqite.Queue
	runner           *jobs.Runner
	dbRO             database.IService
	dbRW             database.IService
	thumbnailService IThumbnailService
}

func NewThumbnailJobRunner(db *sql.DB, dbRO, dbRW database.IService, thumbnailService IThumbnailService) *ThumbnailJobRunner {
	if db == nil {
		panic("db is required")
	}
	if thumbnailService == nil || reflect.ValueOf(thumbnailService).IsNil() {
		panic("implementor of IThumbnailService is required")
	}

	q := goqite.New(goqite.NewOpts{
		DB:   db,
		Name: ThumbnailQueueName,
	})

	runner := jobs.NewRunner(jobs.NewRunnerOpts{
		Limit:        3,
		Log:          slog.Default(),
		PollInterval: 5 * time.Second,
		Queue:        q,
	})

	tjr := &ThumbnailJobRunner{
		queue:            q,
		runner:           runner,
		dbRO:             dbRO,
		dbRW:             dbRW,
		thumbnailService: thumbnailService,
	}

	runner.Register(JobCreateThumbnail, tjr.handleCreateThumbnail)

	return tjr
}

func (tjr *ThumbnailJobRunner) Start(ctx context.Context) {
	logs.Log().Info("[ThumbnailJobRunner] Starting thumbnail job runner")
	tjr.runner.Start(ctx)
}

func (tjr *ThumbnailJobRunner) QueueThumbnailJob(ctx context.Context, params ThumbnailJobParams) error {
	const logtag = "[ThumbnailJobRunner QueueThumbnailJob]"

	tempPayload := ThumbnailJobPayload{ThumbnailJobID: 0}
	payloadBytes, err := json.Marshal(tempPayload)
	if err != nil {
		return errors.Join(errs.ErrJobsCreateFailed, err)
	}

	if err := tjr.queue.Send(ctx, goqite.Message{Body: payloadBytes}); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		return errors.Join(errs.ErrJobsCreateFailed, err)
	}

	msg, err := tjr.queue.Receive(ctx)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		return errors.Join(errs.ErrJobsCreateFailed, err)
	}

	queueID := string(msg.ID)

	if err := tjr.queue.Delete(ctx, msg.ID); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		return errors.Join(errs.ErrJobsCreateFailed, err)
	}

	insertParams := queries.InsertThumbnailJobParams{
		QueueID:    queueID,
		ProductID:  params.ProductID,
		Brand:      params.Brand,
		SourcePath: params.SourcePath,
	}

	thumbnailJob, err := tjr.dbRW.GetQueries().InsertThumbnailJob(ctx, insertParams)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		return errors.Join(errs.ErrJobsCreateFailed, err)
	}

	payload := ThumbnailJobPayload{ThumbnailJobID: thumbnailJob.ID}
	payloadBytes, err = json.Marshal(payload)
	if err != nil {
		return errors.Join(errs.ErrJobsCreateFailed, err)
	}

	if _, err := jobs.Create(ctx, tjr.queue, JobCreateThumbnail, goqite.Message{Body: payloadBytes}); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		return errors.Join(errs.ErrJobsCreateFailed, err)
	}

	logs.LogCtx(ctx).Info(
		logtag,
		zap.Int64("thumbnail_job_id", thumbnailJob.ID),
		zap.String("queue_id", queueID),
		zap.Int64("product_id", params.ProductID),
		zap.String("brand", params.Brand),
	)

	return nil
}

func (tjr *ThumbnailJobRunner) handleCreateThumbnail(ctx context.Context, m []byte) error {
	const logtag = "[ThumbnailJobRunner handleCreateThumbnail]"

	var payload ThumbnailJobPayload
	if err := json.Unmarshal(m, &payload); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		return err
	}

	thumbnailJob, err := tjr.dbRO.GetQueries().GetThumbnailJobByID(ctx, payload.ThumbnailJobID)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Int64("thumbnail_job_id", payload.ThumbnailJobID), zap.Error(err))
		return errors.Join(errs.ErrJobsThumbnailNotFound, err)
	}

	logs.LogCtx(ctx).Info(
		logtag,
		zap.Int64("thumbnail_job_id", thumbnailJob.ID),
		zap.Int64("product_id", thumbnailJob.ProductID),
		zap.String("brand", thumbnailJob.Brand),
		zap.String("source_path", thumbnailJob.SourcePath),
	)

	productImage, err := tjr.dbRO.GetQueries().GetProductImageByProductID(ctx, thumbnailJob.ProductID)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Int64("product_id", thumbnailJob.ProductID), zap.Error(err))
		err2 := tjr.updateJobStatus(ctx, thumbnailJob.ID, "failed", "product image not found")
		return errors.Join(errs.ErrJobsThumbnailFailed, err, err2)
	}

	filename := productImage.Path
	if filename == "" {
		err := errors.New("product has no image path")
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		err2 := tjr.updateJobStatus(ctx, thumbnailJob.ID, "failed", err.Error())
		return errors.Join(errs.ErrJobsThumbnailFailed, err, err2)
	}

	variants, err := tjr.thumbnailService.ProcessImageVariants(ctx, thumbnailJob.SourcePath, thumbnailJob.Brand, filename)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		err2 := tjr.updateJobStatus(ctx, thumbnailJob.ID, "failed", err.Error())
		return errors.Join(errs.ErrJobsThumbnailFailed, err, err2)
	}

	var thumbnailURL string
	for _, v := range variants {
		if v.Size == "640x640" {
			thumbnailURL = v.URL
			break
		}
	}

	if conf.Conf().IsLocal() {
		thumbnailURL = "static/" + thumbnailURL
	}

	if err := tjr.updateProductThumbnail(ctx, thumbnailJob.ProductID, thumbnailURL); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		err2 := tjr.updateJobStatus(ctx, thumbnailJob.ID, "failed", err.Error())
		return errors.Join(errs.ErrJobsThumbnailFailed, err, err2)
	}

	if err := tjr.updateJobStatus(ctx, thumbnailJob.ID, "completed", ""); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		return errors.Join(errs.ErrJobsThumbnailFailed, err)
	}

	logs.LogCtx(ctx).Info(
		logtag,
		zap.String("result", "success"),
		zap.Int64("product_id", thumbnailJob.ProductID),
		zap.String("thumbnail_url", thumbnailURL),
	)

	return nil
}

func (tjr *ThumbnailJobRunner) updateJobStatus(ctx context.Context, jobID int64, status, errorMsg string) error {
	const logtag = "[ThumbnailJobRunner updateJobStatus]"
	if _, err := tjr.dbRW.GetQueries().UpdateThumbnailJobStatus(ctx, queries.UpdateThumbnailJobStatusParams{
		Status:       status,
		ErrorMessage: sql.NullString{String: errorMsg, Valid: errorMsg != ""},
		ID:           jobID,
	}); err != nil {
		logs.Log().Warn(
			logtag,
			zap.Int64("job id", jobID),
			zap.String("status", status),
			zap.String("error message", errorMsg),
			zap.Error(err),
		)
		return err
	}
	return nil
}

func (tjr *ThumbnailJobRunner) updateProductThumbnail(ctx context.Context, productID int64, thumbnailURL string) error {
	existingImg, err := tjr.dbRO.GetQueries().GetProductImageByProductID(ctx, productID)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	_, err = tjr.dbRW.GetQueries().UpdateProductImageThumbnail(ctx, queries.UpdateProductImageThumbnailParams{
		ID:        existingImg.ID,
		Thumbnail: sql.NullString{String: thumbnailURL, Valid: true},
	})
	return err
}
