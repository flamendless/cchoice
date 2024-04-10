-- name: GetProductCategoryByID :one
SELECT *
FROM product_category
WHERE id = ?
LIMIT 1;

-- name: GetProductCategoryByCategory :one
SELECT *
FROM product_category
WHERE category = ?
LIMIT 1;

-- name: GetProductCategoryBySubcategory :one
SELECT *
FROM product_category
WHERE subcategory = ?
LIMIT 1;

-- name: GetProductCategoryByCategoryAndSubcategory :one
SELECT *
FROM product_category
WHERE category = ? AND subcategory = ?
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
