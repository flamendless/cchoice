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
	tbl_products.description as description,
	tbl_products.unit_price_with_vat,
	tbl_products.unit_price_with_vat_currency,
	tbl_product_sales.sale_price_with_vat,
	tbl_product_sales.sale_price_with_vat_currency,
	CASE
		WHEN tbl_product_sales.id IS NOT NULL THEN true
		ELSE false
	END AS is_on_sale,
	tbl_product_sales.discount_type,
	tbl_product_sales.discount_value,
	tbl_brands.name as brand_name,
	COALESCE(
		tbl_product_images.thumbnail,
		'static/images/empty_96x96.webp'
	) AS thumbnail_path,
	tbl_product_specs.weight,
	tbl_product_specs.weight_unit
FROM tbl_checkout_lines
INNER JOIN tbl_products ON tbl_products.id = tbl_checkout_lines.product_id
INNER JOIN tbl_brands ON tbl_brands.id = tbl_products.brand_id
LEFT JOIN tbl_product_images ON tbl_product_images.product_id = tbl_products.id
LEFT JOIN tbl_product_specs ON tbl_product_specs.id = tbl_products.product_specs_id
LEFT JOIN tbl_product_sales
	ON tbl_product_sales.product_id = tbl_products.id
	AND tbl_product_sales.is_active = 1
	AND datetime('now') BETWEEN
		tbl_product_sales.starts_at AND tbl_product_sales.ends_at
WHERE tbl_checkout_lines.checkout_id = ?;

-- name: UpdateCheckoutLineQtyByID :one
UPDATE tbl_checkout_lines
SET quantity = MIN(99, MAX(1, quantity + ?))
WHERE id = ?
RETURNING quantity;

-- name: RemoveItemInCheckoutLinesByID :exec
DELETE FROM tbl_checkout_lines
WHERE checkout_id = ?
	AND id NOT IN (sqlc.slice('ids'))
;

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
	payment_method_type,
	paid_at,
	metadata_remarks,
	metadata_notes,
	metadata_customer_number,
	payment_intent_id
) VALUES (
	?, ?, ?, ?, ?,
	?, ?, ?, ?, ?,
	?, ?, ?, ?, ?
) RETURNING *;

-- name: GetCheckoutPaymentByID :one
SELECT * FROM tbl_checkout_payments
WHERE id = ?
LIMIT 1;

-- name: GetCheckoutPaymentByCheckoutID :one
SELECT * FROM tbl_checkout_payments
WHERE checkout_id = ?
LIMIT 1;

-- name: GetCheckoutPaymentByReferenceNumber :one
SELECT * FROM tbl_checkout_payments
WHERE reference_number = ?
LIMIT 1;

-- name: UpdateCheckoutPaymentOnSuccess :one
UPDATE tbl_checkout_payments
SET status = ?,
	paid_at = DATETIME('now'),
	updated_at = DATETIME('now')
WHERE id = ?
RETURNING *;

-- name: UpdateCheckoutStatus :one
UPDATE tbl_checkouts
SET status = ?,
	updated_at = DATETIME('now')
WHERE id = ?
RETURNING *;
