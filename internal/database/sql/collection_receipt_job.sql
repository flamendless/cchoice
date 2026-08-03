-- name: InsertCollectionReceiptJob :one
INSERT INTO tbl_collection_receipt_jobs (
    queue_id,
    collection_receipt_id,
    job_type,
    staff_id,
    status,
    error_message,
    created_at,
    updated_at
) VALUES (
    ?, ?, ?, ?, ?, ?, datetime('now'), datetime('now')
) RETURNING *;

-- name: GetCollectionReceiptJobByID :one
SELECT sqlc.embed(tbl_collection_receipt_jobs)
FROM tbl_collection_receipt_jobs
WHERE id = ?
LIMIT 1;

-- name: GetLatestCollectionReceiptJobByReceiptID :one
SELECT sqlc.embed(tbl_collection_receipt_jobs)
FROM tbl_collection_receipt_jobs
WHERE collection_receipt_id = ?
ORDER BY id DESC
LIMIT 1;

-- name: GetLatestCollectionReceiptJobByReceiptIDAndType :one
SELECT sqlc.embed(tbl_collection_receipt_jobs)
FROM tbl_collection_receipt_jobs
WHERE collection_receipt_id = ?
    AND job_type = ?
ORDER BY id DESC
LIMIT 1;

-- name: UpdateCollectionReceiptJobStatus :exec
UPDATE tbl_collection_receipt_jobs
SET status = ?, error_message = ?, updated_at = datetime('now')
WHERE id = ?;
