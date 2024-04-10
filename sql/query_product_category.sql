-- name: GetProductCategory :one
SELECT *
FROM product_category
WHERE id = ?
LIMIT 1;

-- name: GetProductCategories :many
SELECT *
FROM product_category
ORDER BY category DESC;

-- name: CreateProductCategory :one
INSERT INTO product_category (
	category,
	subcategory
) VALUES (
	?, ?
) RETURNING *;
