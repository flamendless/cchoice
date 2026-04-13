-- name: CreateResetToken :one
INSERT INTO tbl_password_reset_tokens (
    user_id,
    user_type,
    token_hash,
    expires_at,
    created_at
) VALUES (
    ?,
    ?,
    ?,
    datetime('now', '+15 minutes'),
    datetime('now')
) RETURNING id;

-- name: GetValidResetToken :one
SELECT
    id,
    user_id,
    user_type,
    token_hash,
    expires_at,
    created_at,
    used_at
FROM tbl_password_reset_tokens
WHERE
    token_hash = ?
    AND used_at IS NULL
    AND expires_at > datetime('now')
LIMIT 1;

-- name: GetLatestUnusedResetToken :one
SELECT
    id,
    user_id,
    user_type,
    token_hash,
    expires_at,
    created_at,
    used_at
FROM tbl_password_reset_tokens
WHERE
    user_id = ?
    AND user_type = ?
    AND used_at IS NULL
ORDER BY created_at DESC
LIMIT 1;

-- name: MarkResetTokenAsUsed :exec
UPDATE tbl_password_reset_tokens
SET used_at = datetime('now')
WHERE id = ?;

-- name: InvalidateUserResetTokens :exec
UPDATE tbl_password_reset_tokens
SET used_at = datetime('now')
WHERE user_id = ? AND user_type = ?;
