-- name: GetProductCategoryByID :one
SELECT *
FROM tbl_product_category
WHERE id = ?
LIMIT 1;

-- name: GetProductCategoryByCategory :one
SELECT *
FROM tbl_product_category
WHERE category = ?
LIMIT 1;

-- name: GetProductCategoryByCategoryAndSubcategory :one
SELECT *
FROM tbl_product_category
WHERE category = ? AND subcategory = ?
LIMIT 1;

-- name: CreateProductCategory :one
INSERT INTO tbl_product_category (
	category,
	subcategory
) VALUES (
	?, ?
) RETURNING *;

-- name: CreateProductsCategories :one
INSERT INTO tbl_products_categories (
	product_id,
	category_id
) VALUES (
	?, ?
) RETURNING *;

-- name: SetInitialPromotedProductCategory :many
UPDATE tbl_product_category
SET promoted_at_homepage = true
WHERE
	category LIKE '%grinder%' OR
	category LIKE '%jigsaw%' OR
	category LIKE '%circular-saw%' OR
	category LIKE '%drill%' OR
	category LIKE '%cut-off%' OR
	category LIKE '%mitre-saw%' OR
	category LIKE '%rotary-hammer%' OR
	category LIKE '%demolition-hammer%'
RETURNING id
;

-- name: GetProductCategoriesByPromoted :many
SELECT id, category, subcategory
FROM tbl_product_category
WHERE promoted_at_homepage = ?
LIMIT ?;
