-- name: GetHolidayByDate :one
SELECT
    id,
    date,
    name,
    type,
    created_at,
    updated_at
FROM tbl_holidays
WHERE date = ?
LIMIT 1;

-- name: GetAllHolidays :many
SELECT
    id,
    date,
    name,
    type,
    created_at,
    updated_at
FROM tbl_holidays
ORDER BY date ASC;

-- name: GetHolidaysByDateRange :many
SELECT
    id,
    date,
    name,
    type,
    created_at,
    updated_at
FROM tbl_holidays
WHERE date >= sqlc.arg('start_date') AND date <= sqlc.arg('end_date')
ORDER BY date ASC;

-- name: CreateHoliday :one
INSERT INTO tbl_holidays (
    date,
    name,
    type,
    created_at
) VALUES (
    ?, ?, ?, datetime('now')
) RETURNING id;

-- name: UpdateHoliday :one
UPDATE tbl_holidays
SET
    name = ?,
    type = ?,
    updated_at = datetime('now')
WHERE id = ?
RETURNING id;

-- name: DeleteHoliday :exec
DELETE FROM tbl_holidays WHERE id = ?;
