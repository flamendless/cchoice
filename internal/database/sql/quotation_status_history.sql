-- name: CreateQuotationStatusHistory :one
INSERT INTO tbl_quotation_status_history (
	quotation_id,
	staff_id,
	from_status,
	to_status,
	notes,
	created_at,
	updated_at
) VALUES (
	?, ?, ?, ?, ?,
	datetime('now'),
	datetime('now')
) RETURNING *;

-- name: GetQuotationStatusHistoryByQuotationID :many
SELECT
	h.id,
	h.quotation_id,
	h.staff_id,
	h.from_status,
	h.to_status,
	h.notes,
	h.created_at,
	h.updated_at,
	s.first_name AS staff_first_name,
	s.last_name AS staff_last_name
FROM tbl_quotation_status_history h
LEFT JOIN tbl_staffs s ON s.id = h.staff_id
WHERE h.quotation_id = ?
ORDER BY h.created_at ASC, h.id ASC;
