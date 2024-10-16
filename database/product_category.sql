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

-- name: GetProductsCategoriesByIDs :one
SELECT id FROM tbl_products_categories
WHERE product_id = ? AND category_id = ?
LIMIT 1;

-- name: CreateProductsCategories :one
INSERT INTO tbl_products_categories (
	product_id,
	category_id
) VALUES (?, ?)
ON CONFLICT (product_id, category_id) DO NOTHING
RETURNING *;

-- name: SetInitialPromotedProductCategory :many
UPDATE tbl_product_category
SET promoted_at_homepage = true
WHERE
	category IN (
		'small-angle-grinders',
		'impact-drills',
		'cordless-drill-driver',
		'cut-off-saw',
		'circular-saws',
		'demolition-hammer-hex',
		'demolition-hammer-sds-max'
	)
RETURNING id
;

-- name: GetProductCategoriesByPromoted :many
SELECT
	tbl_product_category.id,
	tbl_product_category.category,
	COUNT(tbl_products_categories.product_id) AS products_count
FROM tbl_product_category
INNER JOIN tbl_products_categories ON tbl_products_categories.category_id = tbl_product_category.id
WHERE promoted_at_homepage = ?
GROUP BY tbl_products_categories.category_id
HAVING tbl_products_categories.product_id
ORDER BY products_count ASC
LIMIT ?;