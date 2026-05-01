-- name: CreateQuotation :one
INSERT INTO tbl_quotations(
	customer_id,
	acknowledged_by_staff_id,
	status,
	created_at,
	updated_at
) VALUES (
	?,
	?,
	'DRAFT',
	datetime('now'),
	datetime('now')
) RETURNING *;

-- name: GetActiveQuotationByCustomerID :one
SELECT * FROM tbl_quotations
WHERE customer_id = ? AND status = 'DRAFT'
ORDER BY created_at DESC
LIMIT 1;

-- name: UpdateQuotationStatus :one
UPDATE tbl_quotations
SET status = ?,
	updated_at = datetime('now')
WHERE id = ?
RETURNING *;

-- name: UpdateQuotationAcknowledgedBy :one
UPDATE tbl_quotations
SET acknowledged_by_staff_id = ?,
	updated_at = datetime('now')
WHERE id = ?
RETURNING *;

-- name: CreateQuotationLine :one
INSERT INTO tbl_quotation_lines(
	quotation_id,
	product_id,
	quantity,
	original_price_snapshot,
	sale_price_snapshot,
	currency,
	created_at,
	updated_at
) VALUES (
	?,
	?,
	?,
	?,
	?,
	?,
	datetime('now'),
	datetime('now')
) RETURNING *;

-- name: GetQuotationLinesByQuotationID :many
SELECT
	tbl_quotation_lines.*,
	tbl_products.serial AS product_serial,
	tbl_products.name AS product_name,
	tbl_brands.name AS brand_name
FROM tbl_quotation_lines
INNER JOIN tbl_products ON tbl_products.id = tbl_quotation_lines.product_id
INNER JOIN tbl_brands ON tbl_brands.id = tbl_products.brand_id
LEFT JOIN tbl_product_sales ON tbl_product_sales.product_id = tbl_products.id AND tbl_product_sales.is_active = 1
WHERE tbl_quotation_lines.quotation_id = ?
ORDER BY tbl_quotation_lines.created_at DESC;

-- name: UpdateQuotationLineQuantity :one
UPDATE tbl_quotation_lines
SET quantity = ?,
	updated_at = datetime('now')
WHERE id = ?
RETURNING *;

-- name: DeleteQuotationLine :exec
DELETE FROM tbl_quotation_lines
WHERE id = ?;

-- name: GetQuotationSummary :one
SELECT
	COUNT(*) AS total_items,
	COALESCE(CAST(SUM(tbl_quotation_lines.quantity * tbl_quotation_lines.original_price_snapshot) AS INTEGER), 0) AS total_original_price,
	COALESCE(CAST(SUM(tbl_quotation_lines.quantity * COALESCE(tbl_quotation_lines.sale_price_snapshot, tbl_quotation_lines.original_price_snapshot)) AS INTEGER), 0) AS total_sale_price,
	CAST(COALESCE(currency, '') AS TEXT) AS currency
FROM tbl_quotation_lines
WHERE tbl_quotation_lines.quotation_id = ?;
