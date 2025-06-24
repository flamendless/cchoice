-- name: GetProductCategoryByID :one
SELECT *
FROM tbl_product_categories
WHERE id = ?
LIMIT 1;

-- name: GetProductCategoryByCategory :one
SELECT *
FROM tbl_product_categories
WHERE category = ?
LIMIT 1;

-- name: GetProductCategoryByCategoryAndSubcategory :one
SELECT *
FROM tbl_product_categories
WHERE category = ? AND subcategory = ?
LIMIT 1;

-- name: CreateProductCategory :one
INSERT INTO tbl_product_categories (
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

-- name: SetInitialPromotedProductCategories :many
UPDATE tbl_product_categories
SET promoted_at_homepage = true
WHERE
	category IN (sqlc.slice('categories'))
RETURNING id
;

-- name: GetProductCategoriesByPromoted :many
SELECT
	tbl_product_categories.id,
	tbl_product_categories.category,
	COUNT(tbl_products_categories.product_id) AS products_count
FROM tbl_product_categories
INNER JOIN tbl_products_categories ON tbl_products_categories.category_id = tbl_product_categories.id
WHERE promoted_at_homepage = ?
GROUP BY tbl_products_categories.category_id
HAVING tbl_products_categories.product_id
ORDER BY tbl_product_categories.category ASC
LIMIT ?;

-- name: GetProductsByCategoryID :many
SELECT
	tbl_products.id,
	tbl_products.name,
	tbl_products.description,
	tbl_products.unit_price_with_vat,
	tbl_products.unit_price_with_vat_currency,
	tbl_brands.name AS brand_name,
	COALESCE(
		tbl_product_images.thumbnail,
		'static/images/empty_96x96.webp'
	) AS thumbnail_path,
	'' as thumbnail_data
FROM tbl_products
INNER JOIN
	tbl_brands ON tbl_brands.id = tbl_products.brand_id
INNER JOIN
	tbl_products_categories ON tbl_products_categories.product_id = tbl_products.id
LEFT JOIN
	tbl_product_images ON tbl_product_images.product_id = tbl_products.id
WHERE tbl_products_categories.category_id = ?
ORDER BY tbl_products.created_at DESC
LIMIT ?
;

-- name: GetProductCategoriesForSections :many
SELECT
	tbl_product_categories.id,
	tbl_product_categories.category,
	tbl_product_categories.subcategory,
	COUNT(tbl_products_categories.product_id) AS products_count
FROM tbl_product_categories
INNER JOIN tbl_products_categories ON tbl_products_categories.category_id = tbl_product_categories.id
GROUP BY tbl_products_categories.category_id
HAVING tbl_products_categories.product_id
ORDER BY tbl_product_categories.category ASC
LIMIT 256;

-- name: GetProductCategoriesForSectionsPagination :many
SELECT
	tbl_product_categories.id,
	tbl_product_categories.category,
	tbl_product_categories.subcategory,
	COUNT(tbl_products_categories.product_id) AS products_count
FROM tbl_product_categories
INNER JOIN tbl_products_categories ON tbl_products_categories.category_id = tbl_product_categories.id
GROUP BY tbl_products_categories.category_id
HAVING tbl_products_categories.product_id
ORDER BY tbl_product_categories.category ASC
LIMIT ?
OFFSET ?
;
