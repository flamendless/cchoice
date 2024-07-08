-- name: GetAuthIDByUserID :one
SELECT id
FROM tbl_auth
WHERE
	user_id = ?
LIMIT 1;

-- name: GetAuthIDAndSecretByUserIDAndUnvalidatedOTP :one
SELECT id, otp_secret
FROM tbl_auth
WHERE
	user_id = ? AND
	otp_enabled = false AND
	otp_secret IS NOT NULL AND otp_secret != '' AND
	recovery_codes IS NOT NULL AND recovery_codes != ''
LIMIT 1;

-- name: CreateAuth :exec
INSERT INTO tbl_auth (
	user_id,
	token,
	otp_enabled
) VALUES (
	?, ?, ?
);

-- name: UpdateAuthTokenByUserID :exec
UPDATE tbl_auth SET token = ? WHERE user_id = ?;

-- name: EnrollOTP :exec
UPDATE tbl_auth SET
	otp_enabled = false,
	otp_secret = ?,
	recovery_codes = ?
WHERE id = ?
;

-- name: ValidateInitialOTP :exec
UPDATE tbl_auth SET
	otp_enabled = true
WHERE
	id = ? AND
	otp_enabled = false AND
	otp_secret IS NOT NULL AND otp_secret != '' AND
	recovery_codes IS NOT NULL AND recovery_codes != ''
;
