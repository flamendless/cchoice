-- name: GetUserByEMailAndUserType :one
SELECT id
FROM tbl_user
WHERE
	email = ? AND
	user_type = ? AND
	status = 'ACTIVE'
LIMIT 1;

-- name: GetUserHashedPassword :one
SELECT password
FROM tbl_user
WHERE
	user_type = 'API' AND
	status = 'ACTIVE' AND
	email = ?
LIMIT 1;

-- name: CreateUser :exec
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
);
