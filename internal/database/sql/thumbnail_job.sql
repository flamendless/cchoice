-- name: InsertThumbnailJob :one
INSERT INTO tbl_thumbnail_jobs (
	queue_id,
	product_id,
	brand,
	source_path,
	created_at,
	updated_at
) VALUES (
	?, ?, ?, ?,
	datetime('now'),
	datetime('now')
) RETURNING *;

-- name: GetThumbnailJobByID :one
SELECT * FROM tbl_thumbnail_jobs
WHERE id = ?
LIMIT 1;

-- name: GetThumbnailJobByProductID :one
SELECT * FROM tbl_thumbnail_jobs
WHERE product_id = ?
LIMIT 1;

-- name: UpdateThumbnailJobStatus :one
UPDATE tbl_thumbnail_jobs
SET status = ?, error_message = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;
