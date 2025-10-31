-- name: CreateOrder :one
INSERT INTO tbl_orders(
	checkout_id,
	checkout_payment_id,
	order_number,
	status,
	customer_name,
	customer_email,
	customer_phone,
	billing_address_line1,
	billing_address_line2,
	billing_city,
	billing_state,
	billing_postal_code,
	billing_country,
	billing_latitude,
	billing_longitude,
	billing_formatted_address,
	billing_place_id,
	shipping_address_line1,
	shipping_address_line2,
	shipping_city,
	shipping_state,
	shipping_postal_code,
	shipping_country,
	shipping_latitude,
	shipping_longitude,
	shipping_formatted_address,
	shipping_place_id,
	subtotal_amount,
	shipping_amount,
	discount_amount,
	total_amount,
	currency,
	shipping_service,
	shipping_order_id,
	shipping_tracking_number,
	notes,
	remarks
) VALUES (
	?, ?, ?, ?, ?,
	?, ?, ?, ?, ?,
	?, ?, ?, ?, ?,
	?, ?, ?, ?, ?,
	?, ?, ?, ?, ?,
	?, ?, ?, ?, ?,
	?, ?, ?, ?, ?,
	?, ?
) RETURNING *;

-- name: GetOrderByID :one
SELECT * FROM tbl_orders
WHERE id = ?
LIMIT 1;

-- name: GetOrderByOrderNumber :one
SELECT * FROM tbl_orders
WHERE order_number = ?
LIMIT 1;

-- name: GetOrderByCheckoutID :one
SELECT * FROM tbl_orders
WHERE checkout_id = ?
LIMIT 1;

-- name: GetOrderByCheckoutPaymentID :one
SELECT * FROM tbl_orders
WHERE checkout_payment_id = ?
LIMIT 1;

-- name: UpdateOrderStatus :one
UPDATE tbl_orders
SET status = ?,
	updated_at = DATETIME('now')
WHERE id = ?
RETURNING *;

-- name: UpdateOrderShippingInfo :one
UPDATE tbl_orders
SET shipping_service = ?,
	shipping_order_id = ?,
	shipping_tracking_number = ?,
	updated_at = DATETIME('now')
WHERE id = ?
RETURNING *;

-- name: CreateOrderLine :one
INSERT INTO tbl_order_lines(
	order_id,
	checkout_line_id,
	product_id,
	name,
	serial,
	description,
	unit_price,
	quantity,
	total_price,
	currency
) VALUES (
	?, ?, ?, ?, ?,
	?, ?, ?, ?, ?
) RETURNING *;

-- name: GetOrderLinesByOrderID :many
SELECT * FROM tbl_order_lines
WHERE order_id = ?
ORDER BY id ASC;

-- name: GetOrderLineByID :one
SELECT * FROM tbl_order_lines
WHERE id = ?
LIMIT 1;

