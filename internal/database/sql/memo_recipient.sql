-- name: CreateMemoRecipient :one
INSERT INTO tbl_memo_recipients (memo_id, staff_id, created_at)
VALUES (?, ?, datetime('now'))
RETURNING id;

-- name: DeleteMemoRecipientsByMemoID :exec
DELETE FROM tbl_memo_recipients
WHERE memo_id = ?;

-- name: GetMemoRecipientStaffIDs :many
SELECT staff_id
FROM tbl_memo_recipients
WHERE memo_id = ?;

-- name: IsMemoRecipient :one
SELECT COUNT(*) AS count
FROM tbl_memo_recipients
WHERE memo_id = ? AND staff_id = ?;

-- name: GetMemoRecipientEmails :many
SELECT
    r.staff_id,
    s.email
FROM tbl_memo_recipients r
JOIN tbl_staffs s ON s.id = r.staff_id
WHERE r.memo_id = ?
AND s.email != ''
AND s.deleted_at = '1970-01-01 00:00:00+00:00'
AND s.status != 'RESIGNED'
ORDER BY s.last_name ASC, s.first_name ASC;

-- name: GetMemoRecipientsWithActions :many
SELECT
    r.staff_id,
    s.first_name,
    s.middle_name,
    s.last_name,
    s.email,
    s.position,
    s.user_type,
    a.status AS action_status,
    a.reject_reason,
    a.accepted_at,
    a.rejected_at,
    a.created_at AS action_created_at
FROM tbl_memo_recipients r
JOIN tbl_staffs s ON s.id = r.staff_id
LEFT JOIN tbl_memo_staff_actions a ON a.memo_id = r.memo_id AND a.staff_id = r.staff_id
WHERE r.memo_id = ?
AND s.deleted_at = '1970-01-01 00:00:00+00:00'
AND s.status != 'RESIGNED'
ORDER BY s.last_name ASC, s.first_name ASC;
