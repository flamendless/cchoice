-- name: GetPromoByID :one
SELECT sqlc.embed(tbl_promos)
FROM tbl_promos
WHERE id = ?
AND deleted_at = '1970-01-01 00:00:00+00:00'
LIMIT 1;

-- name: GetAllPromos :many
SELECT sqlc.embed(tbl_promos)
FROM tbl_promos
WHERE deleted_at = '1970-01-01 00:00:00+00:00'
ORDER BY priority ASC, updated_at DESC;

-- name: GetActivePromos :many
SELECT sqlc.embed(tbl_promos)
FROM tbl_promos
WHERE deleted_at = '1970-01-01 00:00:00+00:00'
AND status = 'PUBLISHED'
AND start_date <= datetime('now')
AND end_date >= datetime('now')
ORDER BY priority ASC;

-- name: CreatePromo :one
INSERT INTO tbl_promos (
    title,
    description,
    media_url,
    start_date,
    end_date,
    type,
    status,
    banner_only,
    priority,
    created_at,
    updated_at,
    deleted_at
) VALUES (
    ?, ?, ?, ?, ?, ?,
    'DRAFT', ?, ?,
    datetime('now'),
    datetime('now'),
    '1970-01-01 00:00:00+00:00'
) RETURNING id;

-- name: UpdatePromo :exec
UPDATE tbl_promos
SET
    title = ?,
    description = ?,
    media_url = ?,
    start_date = ?,
    end_date = ?,
    type = ?,
    status = ?,
    banner_only = ?,
    priority = ?,
    updated_at = datetime('now')
WHERE id = ?
AND deleted_at = '1970-01-01 00:00:00+00:00';

-- name: SoftDeletePromo :exec
UPDATE tbl_promos
SET
    status = 'DELETED',
    deleted_at = datetime('now')
WHERE id = ?
AND deleted_at = '1970-01-01 00:00:00+00:00';
