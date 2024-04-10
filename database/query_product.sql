-- name: GetProduct :one
SELECT *
FROM product
INNER JOIN product_category ON product.product_category_id = product_category.id
WHERE product.id = ?
LIMIT 1;

-- name: GetProducts :many
SELECT *
FROM product
INNER JOIN product_category ON product.product_category_id = product_category.id
ORDER BY created_at DESC;

-- name: CreateProduct :one
INSERT INTO product (
	name,
	description,
	colours,
	sizes,
	segmentation,
	product_category_id,
	unit_price_without_vat,
	unit_price_with_vat,
	unit_price_without_vat_currency,
	unit_price_with_vat_currency,
	status,
	created_at,
	updated_at,
	deleted_at
) VALUES (
	?, ?, ?, ?,
	?, ?, ?, ?,
	?, ?, ?,
	?, ?, ?
) RETURNING *;
