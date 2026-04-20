-- name: GetTrackedLinkByID :one
SELECT
    id,
    name,
    slug,
    destination_url,
    source,
    medium,
    campaign,
    status,
    staff_id,
    created_at,
    updated_at
FROM tbl_tracked_links
WHERE id = ?
LIMIT 1;

-- name: GetTrackedLinkBySlug :one
SELECT
    id,
    name,
    slug,
    destination_url,
    source,
    medium,
    campaign,
    status,
    staff_id,
    created_at,
    updated_at
FROM tbl_tracked_links
WHERE slug = ?
LIMIT 1;

-- name: ListTrackedLinks :many
SELECT
    id,
    name,
    slug,
    destination_url,
    source,
    medium,
    campaign,
    status,
    staff_id,
    created_at,
    updated_at
FROM tbl_tracked_links
WHERE status != 'DELETED'
ORDER BY updated_at DESC;

-- name: CreateTrackedLink :one
INSERT INTO tbl_tracked_links (
    id,
    name,
    slug,
    destination_url,
    source,
    medium,
    campaign,
    status,
    staff_id,
    created_at,
    updated_at
) VALUES (
    ?, ?, ?, ?, ?, ?, ?,
    'DRAFT',
    ?,
    datetime('now'),
    datetime('now')
) RETURNING id;

-- name: UpdateTrackedLink :exec
UPDATE tbl_tracked_links
SET
    name = ?,
    slug = ?,
    destination_url = ?,
    source = ?,
    medium = ?,
    campaign = ?,
    status = ?,
    updated_at = datetime('now')
WHERE id = ?;

-- name: SoftDeleteTrackedLink :exec
UPDATE tbl_tracked_links
SET
    status = 'DELETED',
    updated_at = datetime('now')
WHERE id = ?;
