-- name: CreateStaffAccess :one
INSERT INTO tbl_staff_accesses (
    staff_id,
    login_at,
    user_agent,
    created_at,
    updated_at
) VALUES (
    ?,
    datetime('now'),
    ?,
    datetime('now'),
    datetime('now')
) RETURNING id;

-- name: UpdateStaffAccessLogout :one
UPDATE tbl_staff_accesses
SET
    logout_at = datetime('now'),
    updated_at = datetime('now')
WHERE
    id = ?
RETURNING id;
