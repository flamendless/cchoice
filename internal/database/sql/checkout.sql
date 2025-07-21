-- name: CreateCheckoutLine :one
INSERT INTO tbl_checkout_lines(
	checkout_id,
	product_id,
	name,
	serial,
	description,
	amount,
	currency,
	quantity
) VALUES (
	?, ?, ?, ?,
	?, ?, ?, ?
) RETURNING *;

-- name: GetCheckoutLineByID :one
SELECT
	tbl_checkout_lines.id,
	tbl_checkout_lines.checkout_id,
	tbl_checkout_lines.product_id,
	tbl_checkout_lines.quantity,
	tbl_products.name as name,
	tbl_brands.name as brand_name
FROM tbl_checkout_lines
INNER JOIN tbl_products ON tbl_products.id = tbl_checkout_lines.product_id
INNER JOIN tbl_brands ON tbl_brands.id = tbl_products.brand_id
WHERE tbl_checkout_lines.id = ?
LIMIT 1;

-- name: GetCheckoutLinesByCheckoutID :many
SELECT
	tbl_checkout_lines.id,
	tbl_checkout_lines.checkout_id,
	tbl_checkout_lines.product_id,
	tbl_checkout_lines.quantity,
	tbl_products.name as name,
	tbl_brands.name as brand_name
FROM tbl_checkout_lines
INNER JOIN tbl_products ON tbl_products.id = tbl_checkout_lines.product_id
INNER JOIN tbl_brands ON tbl_brands.id = tbl_products.brand_id
WHERE tbl_checkout_lines.checkout_id = ?;


-- name: CreateCheckout :one
INSERT INTO tbl_checkouts(
	session_id
) VALUES (
	?
) RETURNING *;


-- name: GetCheckoutIDBySessionID :one
SELECT id FROM tbl_checkouts
WHERE session_id = ?;

-- name: CreateCheckoutPayment :one
INSERT INTO tbl_checkout_payments(
	id,
	gateway,
	checkout_id,
	status,
	description,
	total_amount,
	checkout_url,
	client_key,
	reference_number,
	payment_status,
	payment_method_type,
	paid_at,
	metadata_remarks,
	metadata_notes,
	metadata_customer_number
) VALUES (
	?, ?, ?, ?, ?,
	?, ?, ?, ?, ?,
	?, ?, ?, ?, ?
) RETURNING *;
