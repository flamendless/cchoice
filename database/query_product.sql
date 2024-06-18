-- name: GetProductByID :one
SELECT *
FROM tbl_product
INNER JOIN tbl_product_category ON tbl_product.id = tbl_product_category.product_id
INNER JOIN tbl_product_specs ON tbl_product.product_specs_id = tbl_product_specs.id
WHERE tbl_product.id = ?
LIMIT 1;

-- name: GetProductByName :one
SELECT *
FROM tbl_product
INNER JOIN tbl_product_category ON tbl_product.id = tbl_product_category.product_id
INNER JOIN tbl_product_specs ON tbl_product.product_specs_id = tbl_product_specs.id
WHERE tbl_product.name = ?
LIMIT 1;

-- name: GetProductBySerial :one
SELECT *
FROM tbl_product
INNER JOIN tbl_product_category ON tbl_product.id = tbl_product_category.product_id
INNER JOIN tbl_product_specs ON tbl_product.product_specs_id = tbl_product_specs.id
WHERE tbl_product.serial = ?
LIMIT 1;

-- name: GetProducts :many
SELECT *
FROM tbl_product
INNER JOIN tbl_product_category ON tbl_product.id = tbl_product_category.product_id
INNER JOIN tbl_product_specs ON tbl_product.product_specs_id = tbl_product_specs.id
ORDER BY created_at DESC;

-- name: GetProductsByStatus :many
SELECT *
FROM tbl_product
INNER JOIN tbl_product_category ON tbl_product.id = tbl_product_category.product_id
INNER JOIN tbl_product_specs ON tbl_product.product_specs_id = tbl_product_specs.id
WHERE tbl_product.status = ?
ORDER BY created_at DESC;

-- name: GetProductsByStatusSortByNameAsc :many
SELECT *
FROM tbl_product
INNER JOIN tbl_product_category ON tbl_product.id = tbl_product_category.product_id
INNER JOIN tbl_product_specs ON tbl_product.product_specs_id = tbl_product_specs.id
WHERE tbl_product.status = ?
ORDER BY tbl_product.name ASC;

-- name: GetProductsByStatusSortByNameDesc :many
SELECT *
FROM tbl_product
INNER JOIN tbl_product_category ON tbl_product.id = tbl_product_category.product_id
INNER JOIN tbl_product_specs ON tbl_product.product_specs_id = tbl_product_specs.id
WHERE tbl_product.status = ?
ORDER BY tbl_product.name DESC;

-- name: GetProductIDBySerial :one
SELECT id
FROM tbl_product
WHERE tbl_product.serial = ?
LIMIT 1;

-- name: CreateProduct :one
INSERT INTO tbl_product (
	serial,
	name,
	description,
	brand,
	status,
	product_specs_id,
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
	?
) RETURNING *;

-- name: UpdateProduct :execlastid
UPDATE tbl_product
SET
	name = ?,
	description = ?,
	brand = ?,
	status = ?,
	product_specs_id = ?,
	unit_price_without_vat = ?,
	unit_price_with_vat = ?,
	unit_price_without_vat_currency = ?,
	unit_price_with_vat_currency = ?,
	created_at = ?,
	updated_at = ?,
	deleted_at = ?
WHERE id = ?;
