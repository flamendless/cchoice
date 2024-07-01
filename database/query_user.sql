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
