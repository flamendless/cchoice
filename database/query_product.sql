-- name: GetProductByID :one
SELECT *
FROM tbl_product
INNER JOIN tbl_product_category ON tbl_product.id = tbl_product_category.product_id
INNER JOIN tbl_product_specs ON tbl_product.product_specs_id = tbl_product_specs.id
WHERE tbl_product.id = ?
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

-- name: GetProductsByStatusSortByCreationDateAsc :many
SELECT *
FROM tbl_product
INNER JOIN tbl_product_category ON tbl_product.id = tbl_product_category.product_id
INNER JOIN tbl_product_specs ON tbl_product.product_specs_id = tbl_product_specs.id
WHERE tbl_product.status = ?
ORDER BY tbl_product.created_at ASC;

-- name: GetProductsByStatusSortByCreationDateDesc :many
SELECT *
FROM tbl_product
INNER JOIN tbl_product_category ON tbl_product.id = tbl_product_category.product_id
INNER JOIN tbl_product_specs ON tbl_product.product_specs_id = tbl_product_specs.id
WHERE tbl_product.status = ?
ORDER BY tbl_product.created_at DESC;

-- name: GetProductsByFilter :many
SELECT *
FROM tbl_product
INNER JOIN tbl_product_category ON tbl_product.id = tbl_product_category.product_id
INNER JOIN tbl_product_specs ON tbl_product.product_specs_id = tbl_product_specs.id
WHERE
	(tbl_product.status = @status OR @status IS NULL) OR
	(tbl_product.brand = @brand OR @brand IS NULL)
ORDER BY tbl_product.updated_at DESC;

-- name: GetProductsWithSort :many
SELECT *
FROM tbl_product
ORDER BY
	(CASE WHEN @sort = 'sku' AND @dir = 'ASC' THEN  tbl_product.sku END) ASC,
	(CASE WHEN @sort = 'sku' AND @dir = 'DESC' THEN tbl_product.sku END) DESC,
	(CASE WHEN @sort = 'created_at' AND @dir = 'ASC' THEN tbl_product.created_at END) ASC,
	(CASE WHEN @sort = 'created_at' AND @dir = 'DESC' THEN tbl_product.created_at END) DESC
;

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
