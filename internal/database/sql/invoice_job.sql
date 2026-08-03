-- name: InsertInvoiceJob :one
INSERT INTO tbl_invoice_jobs (
    queue_id,
    invoice_id,
    job_type,
    staff_id,
    status,
    error_message,
    created_at,
    updated_at
) VALUES (
    ?, ?, ?, ?, ?, ?, datetime('now'), datetime('now')
) RETURNING *;

-- name: GetInvoiceJobByID :one
SELECT sqlc.embed(tbl_invoice_jobs)
FROM tbl_invoice_jobs
WHERE id = ?
LIMIT 1;

-- name: GetLatestInvoiceJobByInvoiceID :one
SELECT sqlc.embed(tbl_invoice_jobs)
FROM tbl_invoice_jobs
WHERE invoice_id = ?
ORDER BY id DESC
LIMIT 1;

-- name: GetLatestInvoiceJobByInvoiceIDAndType :one
SELECT sqlc.embed(tbl_invoice_jobs)
FROM tbl_invoice_jobs
WHERE invoice_id = ?
    AND job_type = ?
ORDER BY id DESC
LIMIT 1;

-- name: UpdateInvoiceJobStatus :exec
UPDATE tbl_invoice_jobs
SET status = ?, error_message = ?, updated_at = datetime('now')
WHERE id = ?;
