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

-- name: GetMemoRecipientsWithActions :many
SELECT
    r.staff_id,
    s.first_name,
    s.middle_name,
    s.last_name,
    a.status AS action_status,
    a.reject_reason,
    a.accepted_at,
    a.rejected_at,
    a.created_at AS action_created_at
FROM tbl_memo_recipients r
JOIN tbl_staffs s ON s.id = r.staff_id
LEFT JOIN tbl_memo_staff_actions a ON a.memo_id = r.memo_id AND a.staff_id = r.staff_id
WHERE r.memo_id = ?
ORDER BY s.last_name ASC, s.first_name ASC;
