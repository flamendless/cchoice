-- name: GetAuthForEnrollmentByUserID :one
SELECT id
FROM tbl_auth
WHERE
	user_id = ? AND
	otp_enabled = false AND
	otp_status = 'INITIAL'
LIMIT 1;

-- name: GetAuthForOTPValidation :one
SELECT id, otp_secret
FROM tbl_auth
WHERE
	user_id = ? AND
	otp_enabled = true AND
	otp_status = 'INITIAL'
LIMIT 1;

-- name: CreateInitialAuth :exec
INSERT INTO tbl_auth (
	user_id,
	token,
	otp_enabled,
	otp_status
) VALUES (
	?, '', false, 'INITIAL'
);

-- name: UpdateAuthTokenByUserID :exec
UPDATE tbl_auth SET token = ? WHERE user_id = ?;

-- name: EnrollOTP :exec
UPDATE tbl_auth SET
	otp_enabled = true,
	otp_secret = ?,
	recovery_codes = ?,
	otp_status = 'INITIAL'
WHERE id = ?
;

-- name: FinishOTPEnrollment :exec
UPDATE tbl_auth SET
	otp_status = 'ENROLLED'
WHERE
	id = ? AND
	otp_enabled = false AND
	otp_secret IS NOT NULL AND otp_secret != '' AND
	recovery_codes IS NOT NULL AND recovery_codes != ''
;
