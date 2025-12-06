package errs

import "errors"

var (
	ErrJobsQueueInit       = errors.New("[JOBS]: Failed to initialize queue")
	ErrJobsCreateFailed    = errors.New("[JOBS]: Failed to create job")
	ErrJobsEmailNotFound   = errors.New("[JOBS]: Email job not found")
	ErrJobsOrderNotFound   = errors.New("[JOBS]: Order not found for job")
	ErrJobsPaymentNotFound = errors.New("[JOBS]: Payment not found for job")
	ErrJobsSendEmail       = errors.New("[JOBS]: Failed to send email")
)
