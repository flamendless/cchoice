-- name: CreateProductSale :one
INSERT INTO tbl_product_sales (
	product_id,
	sale_price_without_vat,
	sale_price_with_vat,
	sale_price_without_vat_currency,
	sale_price_with_vat_currency,
	discount_type,
	discount_value,
	starts_at,
	ends_at,
	is_active,
	created_at,
	updated_at
) VALUES (
	?, ?, ?, ?, ?,
	?, ?, ?, ?,
	?,
	datetime('now'),
	datetime('now')
) RETURNING *;

-- name: GetActiveSaleByProductID :one
SELECT *
FROM tbl_product_sales
WHERE product_id = ? AND is_active = 1
LIMIT 1;

