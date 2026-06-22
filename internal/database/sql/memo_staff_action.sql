-- name: GetMemoStaffAction :one
SELECT sqlc.embed(tbl_memo_staff_actions)
FROM tbl_memo_staff_actions
WHERE memo_id = ? AND staff_id = ?
LIMIT 1;

-- name: CreateMemoStaffActionAccept :one
INSERT INTO tbl_memo_staff_actions (
    memo_id,
    staff_id,
    status,
    created_at,
    updated_at,
    accepted_at
) VALUES (
    ?, ?, 'ACCEPTED',
    datetime('now'),
    datetime('now'),
    datetime('now')
) RETURNING id;

-- name: CreateMemoStaffActionReject :one
INSERT INTO tbl_memo_staff_actions (
    memo_id,
    staff_id,
    status,
    reject_reason,
    created_at,
    updated_at,
    rejected_at
) VALUES (
    ?, ?, 'REJECTED', ?,
    datetime('now'),
    datetime('now'),
    datetime('now')
) RETURNING id;
