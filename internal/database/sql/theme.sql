-- name: GetThemeByID :one
SELECT sqlc.embed(tbl_themes)
FROM tbl_themes
WHERE id = ?
AND deleted_at = '1970-01-01 00:00:00+00:00'
LIMIT 1;

-- name: SearchThemesSortTitleAsc :many
SELECT sqlc.embed(tbl_themes),
CASE WHEN status = 'PUBLISHED' AND start_date <= date('now') AND end_date >= date('now') THEN 1 ELSE 0 END AS active
FROM tbl_themes
WHERE (deleted_at = '1970-01-01 00:00:00+00:00' OR status = 'DELETED')
AND (@search IS NULL OR @search = '' OR LOWER(title) LIKE '%' || LOWER(@search) || '%')
ORDER BY title ASC;

-- name: SearchThemesSortTitleDesc :many
SELECT sqlc.embed(tbl_themes),
CASE WHEN status = 'PUBLISHED' AND start_date <= date('now') AND end_date >= date('now') THEN 1 ELSE 0 END AS active
FROM tbl_themes
WHERE (deleted_at = '1970-01-01 00:00:00+00:00' OR status = 'DELETED')
AND (@search IS NULL OR @search = '' OR LOWER(title) LIKE '%' || LOWER(@search) || '%')
ORDER BY title DESC;

-- name: SearchThemesSortStartDateAsc :many
SELECT sqlc.embed(tbl_themes),
CASE WHEN status = 'PUBLISHED' AND start_date <= date('now') AND end_date >= date('now') THEN 1 ELSE 0 END AS active
FROM tbl_themes
WHERE (deleted_at = '1970-01-01 00:00:00+00:00' OR status = 'DELETED')
AND (@search IS NULL OR @search = '' OR LOWER(title) LIKE '%' || LOWER(@search) || '%')
ORDER BY start_date ASC;

-- name: SearchThemesSortStartDateDesc :many
SELECT sqlc.embed(tbl_themes),
CASE WHEN status = 'PUBLISHED' AND start_date <= date('now') AND end_date >= date('now') THEN 1 ELSE 0 END AS active
FROM tbl_themes
WHERE (deleted_at = '1970-01-01 00:00:00+00:00' OR status = 'DELETED')
AND (@search IS NULL OR @search = '' OR LOWER(title) LIKE '%' || LOWER(@search) || '%')
ORDER BY start_date DESC;

-- name: SearchThemesSortEndDateAsc :many
SELECT sqlc.embed(tbl_themes),
CASE WHEN status = 'PUBLISHED' AND start_date <= date('now') AND end_date >= date('now') THEN 1 ELSE 0 END AS active
FROM tbl_themes
WHERE (deleted_at = '1970-01-01 00:00:00+00:00' OR status = 'DELETED')
AND (@search IS NULL OR @search = '' OR LOWER(title) LIKE '%' || LOWER(@search) || '%')
ORDER BY end_date ASC;

-- name: SearchThemesSortEndDateDesc :many
SELECT sqlc.embed(tbl_themes),
CASE WHEN status = 'PUBLISHED' AND start_date <= date('now') AND end_date >= date('now') THEN 1 ELSE 0 END AS active
FROM tbl_themes
WHERE (deleted_at = '1970-01-01 00:00:00+00:00' OR status = 'DELETED')
AND (@search IS NULL OR @search = '' OR LOWER(title) LIKE '%' || LOWER(@search) || '%')
ORDER BY end_date DESC;

-- name: SearchThemesSortStatusAsc :many
SELECT sqlc.embed(tbl_themes),
CASE WHEN status = 'PUBLISHED' AND start_date <= date('now') AND end_date >= date('now') THEN 1 ELSE 0 END AS active
FROM tbl_themes
WHERE (deleted_at = '1970-01-01 00:00:00+00:00' OR status = 'DELETED')
AND (@search IS NULL OR @search = '' OR LOWER(title) LIKE '%' || LOWER(@search) || '%')
ORDER BY status ASC;

-- name: SearchThemesSortStatusDesc :many
SELECT sqlc.embed(tbl_themes),
CASE WHEN status = 'PUBLISHED' AND start_date <= date('now') AND end_date >= date('now') THEN 1 ELSE 0 END AS active
FROM tbl_themes
WHERE (deleted_at = '1970-01-01 00:00:00+00:00' OR status = 'DELETED')
AND (@search IS NULL OR @search = '' OR LOWER(title) LIKE '%' || LOWER(@search) || '%')
ORDER BY status DESC;

-- name: GetActiveTheme :one
SELECT sqlc.embed(tbl_themes)
FROM tbl_themes
WHERE deleted_at = '1970-01-01 00:00:00+00:00'
AND status = 'PUBLISHED'
AND start_date <= date('now')
AND end_date >= date('now')
ORDER BY start_date DESC
LIMIT 1;

-- name: GetOverlappingThemes :many
SELECT sqlc.embed(tbl_themes)
FROM tbl_themes
WHERE deleted_at = '1970-01-01 00:00:00+00:00'
AND status != 'DELETED'
AND id != ?
AND start_date <= ?
AND end_date >= ?;

-- name: CreateTheme :one
INSERT INTO tbl_themes (
    title,
    status,
    start_date,
    end_date,
    configuration,
    configuration_type,
    created_by,
    created_at,
    updated_at,
    deleted_at
) VALUES (
    ?, 'DRAFT', ?, ?, ?, ?, ?,
    datetime('now'),
    datetime('now'),
    '1970-01-01 00:00:00+00:00'
) RETURNING id;

-- name: UpdateTheme :exec
UPDATE tbl_themes
SET
    title = ?,
    status = ?,
    start_date = ?,
    end_date = ?,
    configuration = ?,
    configuration_type = ?,
    updated_at = datetime('now')
WHERE id = ?
AND deleted_at = '1970-01-01 00:00:00+00:00';

-- name: SoftDeleteTheme :exec
UPDATE tbl_themes
SET
    status = 'DELETED',
    deleted_at = datetime('now')
WHERE id = ?
AND deleted_at = '1970-01-01 00:00:00+00:00';
