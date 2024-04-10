-- name: GetProductType :one
SELECT *
FROM product_type
WHERE id = ?
LIMIT 1;

-- name: GetProductTypes :many
SELECT *
FROM product_type
ORDER BY id DESC;

-- name: CreateProductType :one
INSERT INTO product_type (
	colours,
	sizes,
	segmentation
) VALUES (
	?, ?, ?
) RETURNING *;
