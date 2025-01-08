-- name: GetSettingsByNames :many
SELECT
	id, name, value
FROM tbl_settings
WHERE
	name IN (sqlc.slice('name'))
LIMIT 100;
