-- name: GetProductCategoryByID :one
SELECT *
FROM tbl_product_category
WHERE id = ?
LIMIT 1;

-- name: GetProductCategoryByProductID :one
SELECT *
FROM tbl_product_category
WHERE product_id = ?
LIMIT 1;

-- name: GetProductCategoryByCategory :one
SELECT *
FROM tbl_product_category
WHERE category = ?
LIMIT 1;

-- name: GetProductCategoryBySubcategory :one
SELECT *
FROM tbl_product_category
WHERE subcategory = ?
LIMIT 1;

-- name: GetProductCategoryByCategoryAndSubcategory :one
SELECT *
FROM tbl_product_category
WHERE category = ? AND subcategory = ?
LIMIT 1;

-- name: GetProductCategories :many
SELECT *
FROM tbl_product_category
ORDER BY category DESC;

-- name: GetProductCategoriesByProductID :many
SELECT *
FROM tbl_product_category
WHERE product_id = ?
ORDER BY id;

-- name: CreateProductCategory :one
INSERT INTO tbl_product_category (
	product_id,
	category,
	subcategory
) VALUES (
	?, ?, ?
) RETURNING *;
