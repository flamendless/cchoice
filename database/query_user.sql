-- name: GetUserByID :one
SELECT *
FROM tbl_user
WHERE
	id = ?
LIMIT 1;

-- name: GetUserWithAuthByID :one
SELECT
	tbl_user.id,
	tbl_auth.otp_enabled
FROM tbl_user
INNER JOIN tbl_auth ON tbl_auth.user_id = tbl_user.id
WHERE
	tbl_user.id = ?
LIMIT 1;

-- name: GetUserEMailAndMobileNoByID :one
SELECT email, mobile_no
FROM tbl_user
WHERE id = ?
LIMIT 1;

-- name: GetUserByEMailAndUserType :one
SELECT id
FROM tbl_user
WHERE
	email = ? AND
	user_type = ? AND
	status = 'ACTIVE'
LIMIT 1;

-- name: GetUserByEMailAndUserTypeAndToken :one
SELECT tbl_user.id
FROM tbl_user
INNER JOIN tbl_auth ON tbl_auth.user_id = tbl_user.id
WHERE
	email = ? AND
	user_type = ? AND
	status = 'ACTIVE' AND
	tbl_auth.token = ?
LIMIT 1;

-- name: GetUserForAuth :one
SELECT
	tbl_user.id,
	tbl_user.password,
	tbl_auth.otp_enabled
FROM tbl_user
INNER JOIN tbl_auth ON tbl_auth.user_id = tbl_user.id
WHERE
	user_type = 'API' AND
	status = 'ACTIVE' AND
	email = ?
LIMIT 1;

-- name: CreateUser :one
INSERT INTO tbl_user (
	first_name,
	middle_name,
	last_name,
	email,
	password,
	mobile_no,
	user_type,
	status
) VALUES (
	?, ?, ?, ?,
	?, ?, ?, ?
) RETURNING id;

-- name: CheckUniqueEMailandMobileNo :one
SELECT
	EXISTS (
		SELECT 1
		FROM tbl_user
		WHERE
			email = ? OR
			mobile_no = ?
		LIMIT 1
	)
;
