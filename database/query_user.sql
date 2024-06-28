-- name: GetUserByEMailAndUserType :one
SELECT id
FROM tbl_user
WHERE
	email = ? AND
	user_type = ?
LIMIT 1;
