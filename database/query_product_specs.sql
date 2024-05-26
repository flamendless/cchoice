-- name: GetProductSpecsByID :one
SELECT *
FROM tbl_product_specs
WHERE id = ?
LIMIT 1;

-- name: GetProductSpecs :many
SELECT *
FROM tbl_product_specs
ORDER BY id DESC;

-- name: CreateProductSpecs :one
INSERT INTO tbl_product_specs (
	colours,
	sizes,
	segmentation,
	part_number,
	power,
	capacity,
	scope_of_supply
) VALUES (
	?, ?, ?, ?,
	?, ?, ?
) RETURNING *;
