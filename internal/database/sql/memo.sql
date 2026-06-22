-- name: GetMemoByID :one
SELECT sqlc.embed(tbl_memos)
FROM tbl_memos
WHERE id = ?
AND deleted_at = '1970-01-01 00:00:00+00:00'
LIMIT 1;

-- name: GetAllMemos :many
SELECT
    sqlc.embed(m),
    s.first_name AS creator_first_name,
    s.middle_name AS creator_middle_name,
    s.last_name AS creator_last_name
FROM tbl_memos m
JOIN tbl_staffs s ON s.id = m.created_by
WHERE m.deleted_at = '1970-01-01 00:00:00+00:00'
ORDER BY m.updated_at DESC;

-- name: GetPendingMemosForStaff :many
SELECT sqlc.embed(m)
FROM tbl_memos m
JOIN tbl_memo_recipients r ON r.memo_id = m.id AND r.staff_id = ?
LEFT JOIN tbl_memo_staff_actions a ON a.memo_id = m.id AND a.staff_id = ?
WHERE m.deleted_at = '1970-01-01 00:00:00+00:00'
AND m.status = 'PUBLISHED'
AND m.start_date <= datetime('now')
AND m.end_date >= datetime('now')
AND a.id IS NULL
ORDER BY m.start_date ASC, m.id ASC;

-- name: CreateMemo :one
INSERT INTO tbl_memos (
    title,
    message,
    file_url,
    status,
    start_date,
    end_date,
    created_by,
    created_at,
    updated_at,
    deleted_at
) VALUES (
    ?, ?, ?, ?, ?, ?, ?,
    datetime('now'),
    datetime('now'),
    '1970-01-01 00:00:00+00:00'
) RETURNING id;

-- name: UpdateMemo :exec
UPDATE tbl_memos
SET
    title = ?,
    message = ?,
    file_url = ?,
    status = ?,
    start_date = ?,
    end_date = ?,
    updated_at = datetime('now')
WHERE id = ?
AND deleted_at = '1970-01-01 00:00:00+00:00';

-- name: SoftDeleteMemo :exec
UPDATE tbl_memos
SET
    status = 'DELETED',
    deleted_at = datetime('now'),
    updated_at = datetime('now')
WHERE id = ?
AND deleted_at = '1970-01-01 00:00:00+00:00';
