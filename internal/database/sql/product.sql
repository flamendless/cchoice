-- name: GetProductsByID :one
SELECT
	*,
	tbl_brands.name AS brand_name
FROM tbl_products
INNER JOIN tbl_product_categories ON tbl_products.id = tbl_product_categories.product_id
INNER JOIN tbl_product_specs ON tbl_products.product_specs_id = tbl_product_specs.id
INNER JOIN tbl_brands ON tbl_brands.id = tbl_products.brand_id
WHERE tbl_products.id = ?
LIMIT 1;

-- name: GetProductsBySerial :one
SELECT
	*,
	tbl_brands.name AS brand_name
FROM tbl_products
INNER JOIN tbl_product_categories ON tbl_products.id = tbl_product_categories.product_id
INNER JOIN tbl_product_specs ON tbl_products.product_specs_id = tbl_product_specs.id
INNER JOIN tbl_brands ON tbl_brands.id = tbl_products.brand_id
WHERE tbl_products.serial = ?
LIMIT 1;

-- name: GetProducts :many
SELECT
	*,
	tbl_brands.name AS brand_name
FROM tbl_products
INNER JOIN tbl_product_categories ON tbl_products.id = tbl_product_categories.product_id
INNER JOIN tbl_product_specs ON tbl_products.product_specs_id = tbl_product_specs.id
INNER JOIN tbl_brands ON tbl_brands.id = tbl_products.brand_id
ORDER BY created_at DESC;

-- name: GetProductsByStatus :many
SELECT
	*,
	tbl_brands.name AS brand_name
FROM tbl_products
INNER JOIN tbl_product_categories ON tbl_products.id = tbl_product_categories.product_id
INNER JOIN tbl_product_specs ON tbl_products.product_specs_id = tbl_product_specs.id
INNER JOIN tbl_brands ON tbl_brands.id = tbl_products.brand_id
WHERE tbl_products.status = ?
ORDER BY created_at DESC;

-- name: GetProductsByStatusSortByNameAsc :many
SELECT
	*,
	tbl_brands.name AS brand_name
FROM tbl_products
INNER JOIN tbl_product_categories ON tbl_products.id = tbl_product_categories.product_id
INNER JOIN tbl_product_specs ON tbl_products.product_specs_id = tbl_product_specs.id
INNER JOIN tbl_brands ON tbl_brands.id = tbl_products.brand_id
WHERE tbl_products.status = ?
ORDER BY LOWER(tbl_products.name) ASC;

-- name: GetProductsByStatusSortByNameDesc :many
SELECT
	*,
	tbl_brands.name AS brand_name
FROM tbl_products
INNER JOIN tbl_product_categories ON tbl_products.id = tbl_product_categories.product_id
INNER JOIN tbl_product_specs ON tbl_products.product_specs_id = tbl_product_specs.id
INNER JOIN tbl_brands ON tbl_brands.id = tbl_products.brand_id
WHERE tbl_products.status = ?
ORDER BY LOWER(tbl_products.name) DESC;

-- name: GetProductsByStatusSortByCreationDateAsc :many
SELECT
	*,
	tbl_brands.name AS brand_name
FROM tbl_products
INNER JOIN tbl_product_categories ON tbl_products.id = tbl_product_categories.product_id
INNER JOIN tbl_product_specs ON tbl_products.product_specs_id = tbl_product_specs.id
INNER JOIN tbl_brands ON tbl_brands.id = tbl_products.brand_id
WHERE tbl_products.status = ?
ORDER BY tbl_products.created_at ASC;

-- name: GetProductsByStatusSortByCreationDateDesc :many
SELECT
	*,
	tbl_brands.name AS brand_name
FROM tbl_products
INNER JOIN tbl_product_categories ON tbl_products.id = tbl_product_categories.product_id
INNER JOIN tbl_product_specs ON tbl_products.product_specs_id = tbl_product_specs.id
INNER JOIN tbl_brands ON tbl_brands.id = tbl_products.brand_id
WHERE tbl_products.status = ?
ORDER BY tbl_products.created_at DESC;

-- name: GetProductsByFilter :many
SELECT
	*,
	tbl_brands.name AS brand_name
FROM tbl_products
INNER JOIN tbl_product_categories ON tbl_products.id = tbl_product_categories.product_id
INNER JOIN tbl_product_specs ON tbl_products.product_specs_id = tbl_product_specs.id
INNER JOIN tbl_brands ON tbl_brands.id = tbl_products.brand_id
WHERE
	(tbl_products.status = @status OR @status IS NULL) OR
	(tbl_brands.name = @brand OR @brand IS NULL)
ORDER BY tbl_products.updated_at DESC;

-- name: GetProductsWithSort :many
SELECT *
FROM tbl_products
ORDER BY
	(CASE WHEN @sort = 'sku' AND @dir = 'ASC' THEN  tbl_products.sku END) ASC,
	(CASE WHEN @sort = 'sku' AND @dir = 'DESC' THEN tbl_products.sku END) DESC,
	(CASE WHEN @sort = 'created_at' AND @dir = 'ASC' THEN tbl_products.created_at END) ASC,
	(CASE WHEN @sort = 'created_at' AND @dir = 'DESC' THEN tbl_products.created_at END) DESC
;

-- name: GetProductIDBySerial :one
SELECT id
FROM tbl_products
WHERE tbl_products.serial = ?
LIMIT 1;

-- name: CreateProducts :one
INSERT INTO tbl_products (
	serial,
	name,
	description,
	brand_id,
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

-- name: UpdateProducts :execlastid
UPDATE tbl_products
SET
	name = ?,
	description = ?,
	brand_id = ?,
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

-- name: GetProductsListing :many
SELECT
	tbl_products.id,
	tbl_products.name,
	tbl_products.description,
	tbl_products.unit_price_with_vat,
	tbl_products.unit_price_with_vat_currency,
	tbl_brands.name AS brand_name
FROM tbl_products
INNER JOIN tbl_brands ON tbl_brands.id = tbl_products.brand_id
ORDER BY tbl_products.created_at DESC
LIMIT ?
;
