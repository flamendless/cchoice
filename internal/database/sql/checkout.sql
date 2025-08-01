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

-- name: DeleteCheckoutLineByID :exec
DELETE FROM tbl_checkout_lines
WHERE id = ?;

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

-- name: CountCheckoutLineByCheckoutID :one
SELECT COUNT(*)
FROM tbl_checkout_lines
WHERE tbl_checkout_lines.checkout_id = ?;

-- name: GetCheckoutLinesByCheckoutID :many
SELECT
	tbl_checkout_lines.id,
	tbl_checkout_lines.checkout_id,
	tbl_checkout_lines.product_id,
	tbl_checkout_lines.quantity,
	tbl_products.name as name,
	tbl_products.unit_price_with_vat,
	tbl_products.unit_price_with_vat_currency,
	tbl_brands.name as brand_name,
	COALESCE(
		tbl_product_images.thumbnail,
		'static/images/empty_96x96.webp'
	) AS thumbnail_path,
	'' as thumbnail_data
FROM tbl_checkout_lines
INNER JOIN tbl_products ON tbl_products.id = tbl_checkout_lines.product_id
INNER JOIN tbl_brands ON tbl_brands.id = tbl_products.brand_id
LEFT JOIN tbl_product_images ON tbl_product_images.product_id = tbl_products.id
WHERE tbl_checkout_lines.checkout_id = ?;

-- name: UpdateCheckoutLineQtyByID :one
UPDATE tbl_checkout_lines SET quantity = quantity + ?
WHERE id = ? AND quantity > 1 AND quantity < 99
RETURNING quantity;

-- name: CheckCheckoutLineExistsByCheckoutIDAndProductID :one
SELECT EXISTS (
	SELECT 1 FROM tbl_checkout_lines
	WHERE checkout_id = ? AND product_id = ?
);

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
