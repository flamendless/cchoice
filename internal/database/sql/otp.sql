-- name: GetOTPEnabledByUserID :one
SELECT otp_enabled
FROM tbl_auth
WHERE user_id = ?
LIMIT 1
;

-- name: GetAuthForOTPValidation :one
SELECT id, otp_secret
FROM tbl_auth
WHERE user_id = ?
LIMIT 1
;

-- name: CreateInitialAuth :exec
INSERT INTO tbl_auth (
	user_id,
	token,
	otp_enabled
) VALUES (
	?,
	'',
	false
);

-- name: EnrollOTP :exec
UPDATE tbl_auth
SET
	otp_secret = ?,
	recovery_codes = ?
WHERE user_id = ?
;

-- name: FinishOTPEnrollment :exec
UPDATE tbl_auth
SET otp_enabled = true
WHERE
	user_id = ? AND
	otp_enabled = false
;
