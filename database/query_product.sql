-- name: GetProductByID :one
SELECT *
FROM tbl_product
INNER JOIN tbl_product_category ON tbl_product.product_category_id = tbl_product_category.id
WHERE tbl_product.id = ?
LIMIT 1;

-- name: GetProductByName :one
SELECT *
FROM tbl_product
INNER JOIN tbl_product_category ON tbl_product.product_category_id = tbl_product_category.id
WHERE tbl_product.name = ?
LIMIT 1;

-- name: GetProductBySerial :one
SELECT *
FROM tbl_product
INNER JOIN tbl_product_category ON tbl_product.product_category_id = tbl_product_category.id
WHERE tbl_product.serial = ?
LIMIT 1;

-- name: GetProducts :many
SELECT *
FROM tbl_product
INNER JOIN tbl_product_category ON tbl_product.product_category_id = tbl_product_category.id
ORDER BY created_at DESC;

-- name: CreateProduct :one
INSERT INTO tbl_product (
	serial,
	name,
	description,
	brand,
	status,
	colours,
	sizes,
	segmentation,
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
	?, ?, ?, ?,
	?, ?, ?, ?,
	?, ?, ?, ?
) RETURNING *;
