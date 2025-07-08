-- name: CreateCheckoutLineItem :execrows
INSERT INTO tbl_checkout_line_items(
	checkout_id,
	amount,
	currency,
	description,
	name,
	quantity
) VALUES (
	?, ?, ?,
	?, ?, ?
) RETURNING *;

-- name: CreateCheckout :one
INSERT INTO tbl_checkouts(
	id,
	gateway,
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
	?, ?, ?, ?
) RETURNING *;
