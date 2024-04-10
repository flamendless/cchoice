-- name: GetProduct :one
SELECT *
FROM product
INNER JOIN product_category ON product.product_category_id = product_category.id
INNER JOIN product_type ON product.product_type_id = product_type.id
WHERE product.id = ?
LIMIT 1;

-- name: GetProducts :many
SELECT *
FROM product
INNER JOIN product_category ON product.product_category_id = product_category.id
INNER JOIN product_type ON product.product_type_id = product_type.id
ORDER BY created_at DESC;

-- name: CreateProduct :one
INSERT INTO product (
	name,
	description,
	product_type_id,
	product_category_id,
	unit_price_without_vat,
	unit_price_with_vat,
	unit_price_without_vat_currency,
	unit_price_with_vat_currency,
	created_at,
	updated_at,
	deleted_at
) VALUES (
	?, ?, ?, ?,
	?, ?,
	?, ?,
	?, ?, ?
) RETURNING *;
