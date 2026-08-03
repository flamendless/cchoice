-- name: InsertDeliveryReceiptJob :one
INSERT INTO tbl_delivery_receipt_jobs (
    queue_id,
    delivery_receipt_id,
    job_type,
    staff_id,
    status,
    error_message,
    created_at,
    updated_at
) VALUES (
    ?, ?, ?, ?, ?, ?, datetime('now'), datetime('now')
) RETURNING *;

-- name: GetDeliveryReceiptJobByID :one
SELECT sqlc.embed(tbl_delivery_receipt_jobs)
FROM tbl_delivery_receipt_jobs
WHERE id = ?
LIMIT 1;

-- name: GetLatestDeliveryReceiptJobByReceiptID :one
SELECT sqlc.embed(tbl_delivery_receipt_jobs)
FROM tbl_delivery_receipt_jobs
WHERE delivery_receipt_id = ?
ORDER BY id DESC
LIMIT 1;

-- name: GetLatestDeliveryReceiptJobByReceiptIDAndType :one
SELECT sqlc.embed(tbl_delivery_receipt_jobs)
FROM tbl_delivery_receipt_jobs
WHERE delivery_receipt_id = ?
    AND job_type = ?
ORDER BY id DESC
LIMIT 1;

-- name: UpdateDeliveryReceiptJobStatus :exec
UPDATE tbl_delivery_receipt_jobs
SET status = ?, error_message = ?, updated_at = datetime('now')
WHERE id = ?;
