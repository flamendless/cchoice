-- name: CreateExternalAPILog :one
INSERT INTO tbl_external_api_logs (
	checkout_id,
	service,
	api,
	endpoint,
	http_method,
	payload,
	response,
	status_code,
	error_message,
	is_successful,
	created_at,
	updated_at
) VALUES (
	?, ?, ?, ?, ?,
	?, ?, ?, ?, ?,
	?, ?
) RETURNING *;

