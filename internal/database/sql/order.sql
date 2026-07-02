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
	shipping_eta,
	notes,
	remarks,
	created_at,
	updated_at
) VALUES (
	?, ?, ?, ?, ?,
	?, ?, ?, ?, ?,
	?, ?, ?, ?, ?,
	?, ?, ?, ?, ?,
	?, ?, ?, ?, ?,
	?, ?, ?, ?, ?,
	?, ?, ?, ?, ?,
	?, ?, ?,
	datetime('now'),
	datetime('now')
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

-- name: UpdateOrderOnPaymentSuccess :one
UPDATE tbl_orders
SET status = ?,
	paid_at = DATETIME('now'),
	updated_at = DATETIME('now')
WHERE id = ?
RETURNING *;

-- name: UpdateOrderShippingInfo :one
UPDATE tbl_orders
SET shipping_service = ?,
	shipping_order_id = ?,
	shipping_tracking_number = ?,
	shipping_eta = ?,
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
	currency,
	created_at,
	updated_at
) VALUES (
	?, ?, ?, ?, ?,
	?, ?, ?, ?, ?,
	datetime('now'),
	datetime('now')
) RETURNING *;

-- name: GetOrderLinesByOrderID :many
SELECT * FROM tbl_order_lines
WHERE order_id = ?
ORDER BY id ASC;

-- name: AdminGetOrderLinesByOrderID :many
SELECT
	tbl_order_lines.id,
	tbl_order_lines.order_id,
	tbl_order_lines.checkout_line_id,
	tbl_order_lines.product_id,
	tbl_order_lines.name,
	tbl_order_lines.serial,
	tbl_order_lines.unit_price,
	tbl_order_lines.quantity,
	tbl_order_lines.total_price,
	tbl_order_lines.currency,
	COALESCE(tbl_product_images.thumbnail, '') AS thumbnail_path,
	tbl_product_images.cdn_url,
	tbl_product_images.cdn_url_thumbnail
FROM tbl_order_lines
LEFT JOIN tbl_product_images ON tbl_product_images.id = (
	SELECT tpi.id
	FROM tbl_product_images tpi
	WHERE tpi.product_id = tbl_order_lines.product_id
	ORDER BY tpi.updated_at DESC
	LIMIT 1
)
WHERE tbl_order_lines.order_id = ?
ORDER BY tbl_order_lines.id ASC;

-- name: GetOrderLineByID :one
SELECT * FROM tbl_order_lines
WHERE id = ?
LIMIT 1;

-- name: AdminCountOrdersForListing :one
SELECT COUNT(*) AS count
FROM tbl_orders
WHERE
	(@search_order_ref IS NULL OR @search_order_ref = '' OR LOWER(order_number) LIKE '%' || LOWER(@search_order_ref) || '%');

-- name: AdminGetOrdersForListingPaginatedUpdatedAtDesc :many
SELECT
	id,
	order_number,
	status,
	paid_at,
	created_at,
	updated_at
FROM tbl_orders
WHERE
	(@search_order_ref IS NULL OR @search_order_ref = '' OR LOWER(order_number) LIKE '%' || LOWER(@search_order_ref) || '%')
ORDER BY updated_at DESC
LIMIT @limit OFFSET @offset;

-- name: AdminGetOrdersForListingPaginatedUpdatedAtAsc :many
SELECT
	id,
	order_number,
	status,
	paid_at,
	created_at,
	updated_at
FROM tbl_orders
WHERE
	(@search_order_ref IS NULL OR @search_order_ref = '' OR LOWER(order_number) LIKE '%' || LOWER(@search_order_ref) || '%')
ORDER BY updated_at ASC
LIMIT @limit OFFSET @offset;

-- name: AdminGetOrdersForListingPaginatedCreatedAtDesc :many
SELECT
	id,
	order_number,
	status,
	paid_at,
	created_at,
	updated_at
FROM tbl_orders
WHERE
	(@search_order_ref IS NULL OR @search_order_ref = '' OR LOWER(order_number) LIKE '%' || LOWER(@search_order_ref) || '%')
ORDER BY created_at DESC
LIMIT @limit OFFSET @offset;

-- name: AdminGetOrdersForListingPaginatedCreatedAtAsc :many
SELECT
	id,
	order_number,
	status,
	paid_at,
	created_at,
	updated_at
FROM tbl_orders
WHERE
	(@search_order_ref IS NULL OR @search_order_ref = '' OR LOWER(order_number) LIKE '%' || LOWER(@search_order_ref) || '%')
ORDER BY created_at ASC
LIMIT @limit OFFSET @offset;

-- name: AdminGetOrdersForListingPaginatedStatusDesc :many
SELECT
	id,
	order_number,
	status,
	paid_at,
	created_at,
	updated_at
FROM tbl_orders
WHERE
	(@search_order_ref IS NULL OR @search_order_ref = '' OR LOWER(order_number) LIKE '%' || LOWER(@search_order_ref) || '%')
ORDER BY status DESC
LIMIT @limit OFFSET @offset;

-- name: AdminGetOrdersForListingPaginatedStatusAsc :many
SELECT
	id,
	order_number,
	status,
	paid_at,
	created_at,
	updated_at
FROM tbl_orders
WHERE
	(@search_order_ref IS NULL OR @search_order_ref = '' OR LOWER(order_number) LIKE '%' || LOWER(@search_order_ref) || '%')
ORDER BY status ASC
LIMIT @limit OFFSET @offset;

-- name: AdminGetOrderDetailsByID :one
SELECT
	o.id,
	o.order_number,
	o.status,
	o.customer_name,
	o.customer_email,
	o.customer_phone,
	o.notes,
	o.remarks,
	o.created_at,
	o.updated_at,
	o.paid_at,
	o.subtotal_amount,
	o.shipping_amount,
	o.discount_amount,
	o.total_amount,
	o.currency,
	o.billing_address_line1,
	o.billing_address_line2,
	o.billing_city,
	o.billing_state,
	o.billing_postal_code,
	o.billing_country,
	o.billing_latitude,
	o.billing_longitude,
	o.billing_formatted_address,
	o.billing_place_id,
	o.shipping_address_line1,
	o.shipping_address_line2,
	o.shipping_city,
	o.shipping_state,
	o.shipping_postal_code,
	o.shipping_country,
	o.shipping_latitude,
	o.shipping_longitude,
	o.shipping_formatted_address,
	o.shipping_place_id,
	o.shipping_service,
	o.shipping_order_id,
	o.shipping_tracking_number,
	o.shipping_eta,
	p.gateway AS payment_gateway,
	p.status AS payment_status,
	p.description AS payment_description,
	p.total_amount AS payment_total_amount,
	p.reference_number AS payment_reference_number,
	p.payment_method_type AS payment_method_type,
	p.paid_at AS payment_paid_at,
	p.metadata_remarks AS payment_metadata_remarks,
	p.metadata_notes AS payment_metadata_notes,
	p.metadata_customer_number AS payment_metadata_customer_number
FROM tbl_orders o
LEFT JOIN tbl_checkout_payments p ON p.id = o.checkout_payment_id
WHERE o.id = ?
LIMIT 1;

