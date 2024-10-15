-- name: UpdateAuthTokenByUserID :exec
UPDATE tbl_auth SET token = ? WHERE user_id = ?;
