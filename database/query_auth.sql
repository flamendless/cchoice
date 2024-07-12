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
	otp_enabled = true
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
	otp_enabled = true
;

-- name: GetAuthOTP :one
SELECT otp_enabled, otp_status
FROM tbl_auth
WHERE user_id = ?
;

-- name: NeedOTP :exec
UPDATE tbl_auth SET
	otp_status = 'SENT_CODE'
WHERE
	user_id = ?
;

-- name: SetOTPStatusValidByUserID :exec
UPDATE tbl_auth SET
	otp_status = 'VALID'
WHERE
	user_id = ?
;
