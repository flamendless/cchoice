-- name: CreateStaffLog :one
INSERT INTO tbl_staff_logs (
    staff_id,
    action,
    module,
    result,
    useragent_id,
    created_at
) VALUES (
    ?,
    ?,
    ?,
    ?,
    ?,
    datetime('now')
) RETURNING id;

-- name: GetStaffLogByID :one
SELECT
    sl.id,
    sl.staff_id,
    sl.created_at,
    sl.action,
    sl.module,
    sl.result,
    sl.useragent_id,
    s.first_name,
    s.middle_name,
    s.last_name
FROM tbl_staff_logs sl
LEFT JOIN tbl_staffs s ON sl.staff_id = s.id
WHERE sl.id = ?
LIMIT 1;

-- name: GetAllStaffLogs :many
SELECT
    sl.id,
    sl.staff_id,
    sl.created_at,
    sl.action,
    sl.module,
    sl.result,
    sl.useragent_id,
    s.first_name,
    s.middle_name,
    s.last_name
FROM tbl_staff_logs sl
LEFT JOIN tbl_staffs s ON sl.staff_id = s.id
ORDER BY sl.created_at DESC;

-- name: GetDistinctStaffLogActions :many
SELECT DISTINCT action
FROM tbl_staff_logs
ORDER BY action;

-- name: GetDistinctStaffLogModules :many
SELECT DISTINCT module
FROM tbl_staff_logs
ORDER BY module;

-- name: GetFilteredStaffLogs :many
SELECT
    sl.id,
    sl.staff_id,
    sl.created_at,
    sl.action,
    sl.module,
    sl.result,
    sl.useragent_id,
    s.first_name,
    s.middle_name,
    s.last_name
FROM tbl_staff_logs sl
LEFT JOIN tbl_staffs s ON sl.staff_id = s.id
WHERE
    (@action = '' OR sl.action = @action)
    AND (@module = '' OR sl.module = @module)
    AND (@staff_id = 0 OR sl.staff_id = @staff_id)
ORDER BY sl.created_at DESC;
