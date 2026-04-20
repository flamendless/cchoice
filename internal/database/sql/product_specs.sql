-- name: GetProductSpecsByID :one
SELECT *
FROM tbl_product_specs
WHERE id = ?
LIMIT 1;

-- name: GetProductSpecs :many
SELECT *
FROM tbl_product_specs
ORDER BY id DESC;

-- name: GetProductSpecsByProductID :one
SELECT tbl_product_specs.*
FROM tbl_product_specs
INNER JOIN tbl_products ON tbl_products.product_specs_id = tbl_product_specs.id
WHERE tbl_products.id = ?
LIMIT 1;

-- name: CreateProductSpecs :one
INSERT INTO tbl_product_specs (
	colours,
	sizes,
	segmentation,
	part_number,
	power,
	capacity,
	scope_of_supply,
	weight,
	weight_unit,
	created_at,
	updated_at
) VALUES (
	?, ?, ?, ?,
	?, ?, ?, ?,
	?,
	datetime('now'),
	datetime('now')
) RETURNING *;

-- name: UpdateProductSpecs :exec
UPDATE tbl_product_specs
SET
	colours = ?,
	sizes = ?,
	segmentation = ?,
	part_number = ?,
	power = ?,
	capacity = ?,
	scope_of_supply = ?,
	weight = ?,
	weight_unit = ?,
	updated_at = datetime('now')
WHERE id = ?;
