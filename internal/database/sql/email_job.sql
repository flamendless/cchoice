-- name: InsertEmailJob :one
INSERT INTO tbl_email_jobs(
	queue_id,
	recipient,
	cc,
	subject,
	template_name,
	order_id,
	checkout_payment_id
) VALUES (
	?, ?, ?, ?, ?, ?, ?
) RETURNING *;

-- name: GetEmailJobByQueueID :one
SELECT * FROM tbl_email_jobs
WHERE queue_id = ?
LIMIT 1;

-- name: GetEmailJobByID :one
SELECT * FROM tbl_email_jobs
WHERE id = ?
LIMIT 1;

-- name: GetEmailJobsByOrderID :many
SELECT * FROM tbl_email_jobs
WHERE order_id = ?
ORDER BY created_at DESC;

-- name: GetEmailJobsByCheckoutPaymentID :many
SELECT * FROM tbl_email_jobs
WHERE checkout_payment_id = ?
ORDER BY created_at DESC;
