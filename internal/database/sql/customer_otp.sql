-- name: CreateOTPCode :one
INSERT INTO tbl_customer_otp_codes (
    customer_id,
    otp_code,
    expires_at,
    created_at
) VALUES (
    ?, ?, datetime('now', '+5 minutes'), datetime('now')
) RETURNING id;

-- name: GetValidOTPCode :one
SELECT
    id,
    customer_id,
    otp_code,
    expires_at,
    created_at,
    used_at
FROM tbl_customer_otp_codes
WHERE
    customer_id = ?
    AND otp_code = ?
    AND used_at IS NULL
    AND expires_at > datetime('now')
LIMIT 1;

-- name: GetLatestUnusedOTPCode :one
SELECT
    id,
    customer_id,
    otp_code,
    expires_at,
    created_at,
    used_at
FROM tbl_customer_otp_codes
WHERE
    customer_id = ?
    AND used_at IS NULL
ORDER BY created_at DESC
LIMIT 1;

-- name: MarkOTPAsUsed :exec
UPDATE tbl_customer_otp_codes
SET used_at = datetime('now')
WHERE id = ?;

-- name: DeleteExpiredOTPs :exec
DELETE FROM tbl_customer_otp_codes
WHERE expires_at <= datetime('now');

-- name: UpdateCustomerStatusToVerified :one
UPDATE tbl_customers
SET
    status = 'VERIFIED',
    updated_at = datetime('now')
WHERE id = ?
RETURNING status;
