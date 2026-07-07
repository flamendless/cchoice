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

-- name: GetQuotationLinesByQuotationIDAndProductID :many
SELECT * FROM tbl_quotation_lines
WHERE quotation_id = ? AND product_id = ?
ORDER BY created_at ASC;

-- name: UpdateQuotationLineQuantity :one
UPDATE tbl_quotation_lines
SET quantity = ?,
	updated_at = datetime('now')
WHERE id = ?
RETURNING *;

-- name: UpdateQuotationLineOnAdd :one
UPDATE tbl_quotation_lines
SET quantity = ?,
	original_price_snapshot = ?,
	sale_price_snapshot = ?,
	currency = ?,
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

-- name: GetQuotationByID :one
SELECT
	q.*,
	c.first_name AS customer_first_name,
	c.middle_name AS customer_middle_name,
	c.last_name AS customer_last_name,
	c.email AS customer_email
FROM tbl_quotations q
INNER JOIN tbl_customers c ON c.id = q.customer_id
WHERE q.id = ?
LIMIT 1;

-- name: ApproveQuotation :one
UPDATE tbl_quotations
SET status = ?,
	acknowledged_by_staff_id = ?,
	updated_at = datetime('now')
WHERE id = ?
RETURNING *;

-- name: AdminCountQuotationsForListing :one
SELECT COUNT(*) AS count
FROM tbl_quotations q
INNER JOIN tbl_customers c ON c.id = q.customer_id
WHERE
	q.status != 'DRAFT'
	AND (@search IS NULL OR @search = '' OR LOWER(c.first_name || ' ' || c.last_name) LIKE '%' || LOWER(@search) || '%' OR LOWER(c.email) LIKE '%' || LOWER(@search) || '%');

-- name: AdminGetQuotationsForListingPaginatedCreatedAtDesc :many
SELECT
	q.id,
	q.status,
	q.created_at,
	q.updated_at,
	q.acknowledged_by_staff_id,
	c.first_name AS customer_first_name,
	c.middle_name AS customer_middle_name,
	c.last_name AS customer_last_name,
	s.first_name AS staff_first_name,
	s.last_name AS staff_last_name,
	COALESCE(line_stats.total_items, 0) AS total_items,
	COALESCE(line_stats.total_original_price, 0) AS total_original_price,
	COALESCE(line_stats.total_sale_price, 0) AS total_sale_price,
	COALESCE(line_stats.currency, '') AS currency
FROM tbl_quotations q
INNER JOIN tbl_customers c ON c.id = q.customer_id
LEFT JOIN tbl_staffs s ON s.id = q.acknowledged_by_staff_id
LEFT JOIN (
	SELECT
		quotation_id,
		COUNT(*) AS total_items,
		CAST(SUM(quantity * original_price_snapshot) AS INTEGER) AS total_original_price,
		CAST(SUM(quantity * COALESCE(sale_price_snapshot, original_price_snapshot)) AS INTEGER) AS total_sale_price,
		MAX(currency) AS currency
	FROM tbl_quotation_lines
	GROUP BY quotation_id
) line_stats ON line_stats.quotation_id = q.id
WHERE
	q.status != 'DRAFT'
	AND (@search IS NULL OR @search = '' OR LOWER(c.first_name || ' ' || c.last_name) LIKE '%' || LOWER(@search) || '%' OR LOWER(c.email) LIKE '%' || LOWER(@search) || '%')
ORDER BY q.created_at DESC
LIMIT @limit OFFSET @offset;

-- name: AdminGetQuotationsForListingPaginatedCreatedAtAsc :many
SELECT
	q.id,
	q.status,
	q.created_at,
	q.updated_at,
	q.acknowledged_by_staff_id,
	c.first_name AS customer_first_name,
	c.middle_name AS customer_middle_name,
	c.last_name AS customer_last_name,
	s.first_name AS staff_first_name,
	s.last_name AS staff_last_name,
	COALESCE(line_stats.total_items, 0) AS total_items,
	COALESCE(line_stats.total_original_price, 0) AS total_original_price,
	COALESCE(line_stats.total_sale_price, 0) AS total_sale_price,
	COALESCE(line_stats.currency, '') AS currency
FROM tbl_quotations q
INNER JOIN tbl_customers c ON c.id = q.customer_id
LEFT JOIN tbl_staffs s ON s.id = q.acknowledged_by_staff_id
LEFT JOIN (
	SELECT
		quotation_id,
		COUNT(*) AS total_items,
		CAST(SUM(quantity * original_price_snapshot) AS INTEGER) AS total_original_price,
		CAST(SUM(quantity * COALESCE(sale_price_snapshot, original_price_snapshot)) AS INTEGER) AS total_sale_price,
		MAX(currency) AS currency
	FROM tbl_quotation_lines
	GROUP BY quotation_id
) line_stats ON line_stats.quotation_id = q.id
WHERE
	q.status != 'DRAFT'
	AND (@search IS NULL OR @search = '' OR LOWER(c.first_name || ' ' || c.last_name) LIKE '%' || LOWER(@search) || '%' OR LOWER(c.email) LIKE '%' || LOWER(@search) || '%')
ORDER BY q.created_at ASC
LIMIT @limit OFFSET @offset;

-- name: AdminGetQuotationsForListingPaginatedStatusDesc :many
SELECT
	q.id,
	q.status,
	q.created_at,
	q.updated_at,
	q.acknowledged_by_staff_id,
	c.first_name AS customer_first_name,
	c.middle_name AS customer_middle_name,
	c.last_name AS customer_last_name,
	s.first_name AS staff_first_name,
	s.last_name AS staff_last_name,
	COALESCE(line_stats.total_items, 0) AS total_items,
	COALESCE(line_stats.total_original_price, 0) AS total_original_price,
	COALESCE(line_stats.total_sale_price, 0) AS total_sale_price,
	COALESCE(line_stats.currency, '') AS currency
FROM tbl_quotations q
INNER JOIN tbl_customers c ON c.id = q.customer_id
LEFT JOIN tbl_staffs s ON s.id = q.acknowledged_by_staff_id
LEFT JOIN (
	SELECT
		quotation_id,
		COUNT(*) AS total_items,
		CAST(SUM(quantity * original_price_snapshot) AS INTEGER) AS total_original_price,
		CAST(SUM(quantity * COALESCE(sale_price_snapshot, original_price_snapshot)) AS INTEGER) AS total_sale_price,
		MAX(currency) AS currency
	FROM tbl_quotation_lines
	GROUP BY quotation_id
) line_stats ON line_stats.quotation_id = q.id
WHERE
	q.status != 'DRAFT'
	AND (@search IS NULL OR @search = '' OR LOWER(c.first_name || ' ' || c.last_name) LIKE '%' || LOWER(@search) || '%' OR LOWER(c.email) LIKE '%' || LOWER(@search) || '%')
ORDER BY q.status DESC
LIMIT @limit OFFSET @offset;

-- name: AdminGetQuotationsForListingPaginatedStatusAsc :many
SELECT
	q.id,
	q.status,
	q.created_at,
	q.updated_at,
	q.acknowledged_by_staff_id,
	c.first_name AS customer_first_name,
	c.middle_name AS customer_middle_name,
	c.last_name AS customer_last_name,
	s.first_name AS staff_first_name,
	s.last_name AS staff_last_name,
	COALESCE(line_stats.total_items, 0) AS total_items,
	COALESCE(line_stats.total_original_price, 0) AS total_original_price,
	COALESCE(line_stats.total_sale_price, 0) AS total_sale_price,
	COALESCE(line_stats.currency, '') AS currency
FROM tbl_quotations q
INNER JOIN tbl_customers c ON c.id = q.customer_id
LEFT JOIN tbl_staffs s ON s.id = q.acknowledged_by_staff_id
LEFT JOIN (
	SELECT
		quotation_id,
		COUNT(*) AS total_items,
		CAST(SUM(quantity * original_price_snapshot) AS INTEGER) AS total_original_price,
		CAST(SUM(quantity * COALESCE(sale_price_snapshot, original_price_snapshot)) AS INTEGER) AS total_sale_price,
		MAX(currency) AS currency
	FROM tbl_quotation_lines
	GROUP BY quotation_id
) line_stats ON line_stats.quotation_id = q.id
WHERE
	q.status != 'DRAFT'
	AND (@search IS NULL OR @search = '' OR LOWER(c.first_name || ' ' || c.last_name) LIKE '%' || LOWER(@search) || '%' OR LOWER(c.email) LIKE '%' || LOWER(@search) || '%')
ORDER BY q.status ASC
LIMIT @limit OFFSET @offset;

-- name: CustomerCountQuotationsForListing :one
SELECT COUNT(*) AS count
FROM tbl_quotations q
WHERE
	q.customer_id = @customer_id
	AND q.status != 'DRAFT';

-- name: CustomerGetQuotationsForListingPaginatedCreatedAtDesc :many
SELECT
	q.id,
	q.status,
	q.created_at,
	q.updated_at,
	COALESCE(line_stats.total_items, 0) AS total_items,
	COALESCE(line_stats.total_original_price, 0) AS total_original_price,
	COALESCE(line_stats.total_sale_price, 0) AS total_sale_price,
	COALESCE(line_stats.currency, '') AS currency
FROM tbl_quotations q
LEFT JOIN (
	SELECT
		quotation_id,
		COUNT(*) AS total_items,
		CAST(SUM(quantity * original_price_snapshot) AS INTEGER) AS total_original_price,
		CAST(SUM(quantity * COALESCE(sale_price_snapshot, original_price_snapshot)) AS INTEGER) AS total_sale_price,
		MAX(currency) AS currency
	FROM tbl_quotation_lines
	GROUP BY quotation_id
) line_stats ON line_stats.quotation_id = q.id
WHERE
	q.customer_id = @customer_id
	AND q.status != 'DRAFT'
ORDER BY q.created_at DESC
LIMIT @limit OFFSET @offset;

-- name: CustomerGetQuotationsForListingPaginatedCreatedAtAsc :many
SELECT
	q.id,
	q.status,
	q.created_at,
	q.updated_at,
	COALESCE(line_stats.total_items, 0) AS total_items,
	COALESCE(line_stats.total_original_price, 0) AS total_original_price,
	COALESCE(line_stats.total_sale_price, 0) AS total_sale_price,
	COALESCE(line_stats.currency, '') AS currency
FROM tbl_quotations q
LEFT JOIN (
	SELECT
		quotation_id,
		COUNT(*) AS total_items,
		CAST(SUM(quantity * original_price_snapshot) AS INTEGER) AS total_original_price,
		CAST(SUM(quantity * COALESCE(sale_price_snapshot, original_price_snapshot)) AS INTEGER) AS total_sale_price,
		MAX(currency) AS currency
	FROM tbl_quotation_lines
	GROUP BY quotation_id
) line_stats ON line_stats.quotation_id = q.id
WHERE
	q.customer_id = @customer_id
	AND q.status != 'DRAFT'
ORDER BY q.created_at ASC
LIMIT @limit OFFSET @offset;

-- name: CustomerGetQuotationsForListingPaginatedStatusDesc :many
SELECT
	q.id,
	q.status,
	q.created_at,
	q.updated_at,
	COALESCE(line_stats.total_items, 0) AS total_items,
	COALESCE(line_stats.total_original_price, 0) AS total_original_price,
	COALESCE(line_stats.total_sale_price, 0) AS total_sale_price,
	COALESCE(line_stats.currency, '') AS currency
FROM tbl_quotations q
LEFT JOIN (
	SELECT
		quotation_id,
		COUNT(*) AS total_items,
		CAST(SUM(quantity * original_price_snapshot) AS INTEGER) AS total_original_price,
		CAST(SUM(quantity * COALESCE(sale_price_snapshot, original_price_snapshot)) AS INTEGER) AS total_sale_price,
		MAX(currency) AS currency
	FROM tbl_quotation_lines
	GROUP BY quotation_id
) line_stats ON line_stats.quotation_id = q.id
WHERE
	q.customer_id = @customer_id
	AND q.status != 'DRAFT'
ORDER BY q.status DESC
LIMIT @limit OFFSET @offset;

-- name: CustomerGetQuotationsForListingPaginatedStatusAsc :many
SELECT
	q.id,
	q.status,
	q.created_at,
	q.updated_at,
	COALESCE(line_stats.total_items, 0) AS total_items,
	COALESCE(line_stats.total_original_price, 0) AS total_original_price,
	COALESCE(line_stats.total_sale_price, 0) AS total_sale_price,
	COALESCE(line_stats.currency, '') AS currency
FROM tbl_quotations q
LEFT JOIN (
	SELECT
		quotation_id,
		COUNT(*) AS total_items,
		CAST(SUM(quantity * original_price_snapshot) AS INTEGER) AS total_original_price,
		CAST(SUM(quantity * COALESCE(sale_price_snapshot, original_price_snapshot)) AS INTEGER) AS total_sale_price,
		MAX(currency) AS currency
	FROM tbl_quotation_lines
	GROUP BY quotation_id
) line_stats ON line_stats.quotation_id = q.id
WHERE
	q.customer_id = @customer_id
	AND q.status != 'DRAFT'
ORDER BY q.status ASC
LIMIT @limit OFFSET @offset;
