-- name: GetSettingsByNames :many
SELECT
	id, name, value
FROM tbl_settings
WHERE
	name IN (sqlc.slice('name'))
LIMIT 100;

-- name: GetSettingsCOD :one
SELECT CAST(
		CASE
		WHEN value = 'true' then true
		ELSE false
		END
		AS BOOLEAN
	)
FROM tbl_settings
WHERE name = 'cash_on_delivery'
LIMIT 1
;
