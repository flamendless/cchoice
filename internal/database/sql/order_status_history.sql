-- name: CreateOrderStatusHistory :one
INSERT INTO tbl_order_status_history (
    order_id,
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

-- name: GetOrderStatusHistoryByOrderID :many
SELECT
    h.id,
    h.order_id,
    h.staff_id,
    h.from_status,
    h.to_status,
    h.notes,
    h.created_at,
    h.updated_at,
    s.first_name AS staff_first_name,
    s.last_name AS staff_last_name
FROM tbl_order_status_history h
LEFT JOIN tbl_staffs s ON s.id = h.staff_id
WHERE h.order_id = ?
ORDER BY h.created_at ASC, h.id ASC;
