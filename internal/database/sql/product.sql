-- name: GetProductsByID :one
SELECT
	tbl_products.*,
	tbl_brands.name AS brand_name,
	COALESCE(pc.category, '') AS product_category,
	COALESCE(pc.subcategory, '') AS product_subcategory,
	tbl_product_specs.*
FROM tbl_products
INNER JOIN tbl_product_specs ON tbl_products.product_specs_id = tbl_product_specs.id
INNER JOIN tbl_brands ON tbl_brands.id = tbl_products.brand_id
LEFT JOIN tbl_products_categories ON tbl_products_categories.product_id = tbl_products.id
LEFT JOIN tbl_product_categories AS pc ON pc.id = tbl_products_categories.category_id
WHERE tbl_products.id = ?
LIMIT 1;

-- name: ValidateUniqueSerial :one
SELECT
	id
FROM tbl_products
WHERE tbl_products.serial = ?
LIMIT 1;

-- name: GetProductByName :one
SELECT
	*
	-- tbl_brands.name AS brand_name
FROM tbl_products
-- INNER JOIN tbl_product_categories ON tbl_products.id = tbl_product_categories.product_id
-- INNER JOIN tbl_product_specs ON tbl_products.product_specs_id = tbl_product_specs.id
-- INNER JOIN tbl_brands ON tbl_brands.id = tbl_products.brand_id
WHERE tbl_products.name = ?
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

-- name: CheckProductExistsByID :one
SELECT 1
FROM tbl_products
WHERE tbl_products.id = ?
LIMIT 1;

-- name: CreateProducts :one
INSERT INTO tbl_products (
	serial,
	name,
	slug,
	description,
	brand_id,
	status,
	product_specs_id,
	unit_price_without_vat,
	unit_price_with_vat,
	unit_price_without_vat_currency,
	unit_price_with_vat_currency
) VALUES (
	?, ?, ?, ?,
	?, ?, ?, ?,
	?, ?, ?
) RETURNING *;

-- name: UpdateProducts :execlastid
UPDATE tbl_products
SET
	name = ?,
	slug = ?,
	description = ?,
	brand_id = ?,
	status = ?,
	product_specs_id = ?,
	unit_price_without_vat = ?,
	unit_price_with_vat = ?,
	unit_price_without_vat_currency = ?,
	unit_price_with_vat_currency = ?,
	updated_at = datetime('now')
WHERE id = ?;

-- name: AdminGetProductsForListing :many
SELECT
	tbl_products.id,
	tbl_products.name,
	tbl_products.serial,
	tbl_products.slug,
	tbl_products.description,
	tbl_brands.name AS brand_name,
	tbl_products.status,
	tbl_products.unit_price_with_vat,
	tbl_products.unit_price_with_vat_currency,
	tbl_products.created_at,
	tbl_products.updated_at,
	COALESCE(tbl_product_images.thumbnail, '') AS thumbnail_path,
	tbl_product_images.cdn_url,
	tbl_product_images.cdn_url_thumbnail,
	COALESCE(tbl_product_specs.colours, '') AS colours,
	COALESCE(tbl_product_specs.sizes, '') AS sizes,
	COALESCE(tbl_product_specs.segmentation, '') AS segmentation,
	COALESCE(tbl_product_specs.part_number, '') AS part_number,
	COALESCE(tbl_product_specs.power, '') AS power,
	COALESCE(tbl_product_specs.capacity, '') AS capacity,
	COALESCE(tbl_product_specs.scope_of_supply, '') AS scope_of_supply,
	COALESCE(tbl_product_specs.weight, 0) AS weight,
	COALESCE(tbl_product_specs.weight_unit, '') AS weight_unit,
	COALESCE(categories.category, '') AS category,
	COALESCE(categories.subcategory, '') AS subcategory
FROM tbl_products
INNER JOIN tbl_brands ON tbl_brands.id = tbl_products.brand_id
LEFT JOIN tbl_product_images ON tbl_product_images.product_id = tbl_products.id
LEFT JOIN tbl_product_specs ON tbl_product_specs.id = tbl_products.product_specs_id
LEFT JOIN (
	SELECT
		tbl_products_categories.product_id,
		GROUP_CONCAT(tbl_product_categories.category, ', ') AS category,
		GROUP_CONCAT(tbl_product_categories.subcategory, ', ') AS subcategory
	FROM tbl_products_categories
	INNER JOIN tbl_product_categories ON tbl_product_categories.id = tbl_products_categories.category_id
	GROUP BY tbl_products_categories.product_id
) AS categories ON categories.product_id = tbl_products.id
WHERE
	(@search IS NULL OR @search = '' OR LOWER(tbl_products.serial) LIKE '%' || LOWER(@search) || '%')
	AND (@status IS NULL OR @status = '' OR tbl_products.status = @status)
ORDER BY
	CASE tbl_products.status
		WHEN 'DRAFT' THEN 1
		WHEN 'ACTIVE' THEN 2
		WHEN 'DELETED' THEN 3
		ELSE 4
	END,
	tbl_products.updated_at DESC
;


--TODO: (Brandon) if sqlc releases PR #3498
--      replace WHERE with `tbl_products_fts MATCH ?`
-- name: GetProductsBySearchQuery :many
SELECT
	tbl_products.id,
	tbl_products.name,
	tbl_products.unit_price_with_vat,
	tbl_products.unit_price_with_vat_currency,
	tbl_brands.name AS brand_name,
	COALESCE(
		tbl_product_images.thumbnail,
		'static/images/empty_96x96.webp'
	) AS thumbnail_path,
	tbl_product_images.cdn_url,
	tbl_product_images.cdn_url_thumbnail
FROM tbl_products_fts
INNER JOIN tbl_products ON tbl_products.id = tbl_products_fts.rowid
INNER JOIN tbl_brands ON tbl_brands.id = tbl_products.brand_id
LEFT JOIN tbl_product_images ON tbl_product_images.product_id = tbl_products.id
WHERE
	tbl_products.status = 'ACTIVE'
	AND thumbnail_path != 'static/images/empty_96x96.webp'
	AND tbl_products_fts.name MATCH ?
LIMIT ?;

-- name: GetRandomProductOnSale :one
SELECT
	tbl_products.id,
	tbl_products.name,
	tbl_products.description,
	tbl_products.unit_price_with_vat,
	tbl_products.unit_price_with_vat_currency,
	tbl_product_sales.sale_price_with_vat,
	tbl_product_sales.sale_price_with_vat_currency,
	CASE
		WHEN tbl_product_sales.id IS NOT NULL THEN true
		ELSE false
	END AS is_on_sale,
	tbl_product_sales.discount_type,
	tbl_product_sales.discount_value,
	tbl_brands.name AS brand_name,
	COALESCE(
		tbl_product_images.thumbnail,
		'static/images/empty_96x96.webp'
	) AS thumbnail_path,
	tbl_product_images.cdn_url,
	tbl_product_images.cdn_url_thumbnail
FROM tbl_products
INNER JOIN tbl_brands ON tbl_brands.id = tbl_products.brand_id
LEFT JOIN tbl_product_sales
	ON tbl_product_sales.product_id = tbl_products.id
	AND tbl_product_sales.is_active = 1
	AND datetime('now') BETWEEN
		tbl_product_sales.starts_at AND tbl_product_sales.ends_at
LEFT JOIN tbl_product_images ON tbl_product_images.product_id = tbl_products.id
WHERE
	tbl_products.status = 'ACTIVE'
	AND thumbnail_path != 'static/images/empty_96x96.webp'
	AND tbl_product_sales.id IS NOT NULL
ORDER BY RANDOM()
LIMIT 1;

-- name: SoftDeleteProduct :exec
UPDATE tbl_products
SET status = 'DELETED', updated_at = datetime('now'), deleted_at = datetime('now')
WHERE id = ?;

-- name: UpdateProductsStatus :exec
UPDATE tbl_products
SET status = ?, updated_at = datetime('now')
WHERE id = ?;

-- name: GetProductPage :one
SELECT
	tbl_products.id,
	tbl_products.serial,
	tbl_products.name,
	tbl_products.description,
	tbl_products.status,
	tbl_products.unit_price_without_vat,
	tbl_products.unit_price_without_vat_currency,
	tbl_products.unit_price_with_vat,
	tbl_products.unit_price_with_vat_currency,
	tbl_products.created_at,
	tbl_products.updated_at,
	tbl_brands.id AS brand_id,
	tbl_brands.name AS brand_name,
	tbl_brand_images.path AS brand_thumbnail_path,
	tbl_brand_images.s3_url AS brand_thumbnail_url,
	COALESCE(pc.category, '') AS product_category,
	COALESCE(pc.subcategory, '') AS product_subcategory,
	COALESCE(tbl_product_images.path, '') AS image_path,
	COALESCE(tbl_product_images.thumbnail, '') AS thumbnail_path,
	COALESCE(tbl_product_images.cdn_url, '') AS cdn_url,
	COALESCE(tbl_product_images.cdn_url_thumbnail, '') AS cdn_url_thumbnail,
	COALESCE(tbl_product_specs.colours, '') AS colours,
	COALESCE(tbl_product_specs.sizes, '') AS sizes,
	COALESCE(tbl_product_specs.segmentation, '') AS segmentation,
	COALESCE(tbl_product_specs.part_number, '') AS part_number,
	COALESCE(tbl_product_specs.power, '') AS power,
	COALESCE(tbl_product_specs.capacity, '') AS capacity,
	COALESCE(tbl_product_specs.scope_of_supply, '') AS scope_of_supply,
	COALESCE(tbl_product_specs.weight, 0) AS weight,
	COALESCE(tbl_product_specs.weight_unit, '') AS weight_unit,
	CASE
		WHEN tbl_product_sales.id IS NOT NULL THEN true
		ELSE false
	END AS is_on_sale,
	CASE
		WHEN tbl_product_sales.id IS NOT NULL THEN tbl_product_sales.sale_price_with_vat
		ELSE tbl_products.unit_price_with_vat
	END AS sale_price_with_vat,
	CASE
		WHEN tbl_product_sales.id IS NOT NULL THEN tbl_product_sales.sale_price_with_vat_currency
		ELSE tbl_products.unit_price_with_vat_currency
	END AS sale_price_with_vat_currency,
	CASE
		WHEN tbl_product_sales.id IS NOT NULL THEN tbl_product_sales.discount_type
		ELSE ''
	END AS discount_type,
	CASE
		WHEN tbl_product_sales.id IS NOT NULL THEN tbl_product_sales.discount_value
		ELSE 0
	END AS discount_value
FROM tbl_products
INNER JOIN tbl_brands ON tbl_brands.id = tbl_products.brand_id
INNER JOIN tbl_brand_images ON tbl_brand_images.brand_id = tbl_brands.id
LEFT JOIN tbl_product_specs ON tbl_product_specs.id = tbl_products.product_specs_id
LEFT JOIN tbl_product_images ON tbl_product_images.product_id = tbl_products.id
LEFT JOIN tbl_products_categories ON tbl_products_categories.product_id = tbl_products.id
LEFT JOIN tbl_product_categories AS pc ON pc.id = tbl_products_categories.category_id
LEFT JOIN tbl_product_sales
	ON tbl_product_sales.product_id = tbl_products.id
	AND tbl_product_sales.is_active = 1
	AND datetime('now') BETWEEN tbl_product_sales.starts_at AND tbl_product_sales.ends_at
WHERE tbl_products.slug = ? AND tbl_products.status = 'ACTIVE'
LIMIT 1;

-- name: GetRelatedProductsByCategory :many
SELECT
	tbl_products.id,
	tbl_products.name,
	tbl_products.serial,
	tbl_products.unit_price_with_vat,
	tbl_products.unit_price_with_vat_currency,
	tbl_brands.name AS brand_name,
	COALESCE(tbl_product_images.thumbnail, '') AS thumbnail_path,
	COALESCE(tbl_product_images.cdn_url, '') AS cdn_url,
	COALESCE(tbl_product_images.cdn_url_thumbnail, '') AS cdn_url_thumbnail,
	CASE
		WHEN tbl_product_sales.id IS NOT NULL THEN true
		ELSE false
	END AS is_on_sale,
	CASE
		WHEN tbl_product_sales.id IS NOT NULL THEN tbl_product_sales.sale_price_with_vat
		ELSE tbl_products.unit_price_with_vat
	END AS sale_price_with_vat,
	CASE
		WHEN tbl_product_sales.id IS NOT NULL THEN tbl_product_sales.sale_price_with_vat_currency
		ELSE tbl_products.unit_price_with_vat_currency
	END AS sale_price_with_vat_currency
FROM tbl_products
INNER JOIN tbl_brands ON tbl_brands.id = tbl_products.brand_id
LEFT JOIN tbl_product_images ON tbl_product_images.product_id = tbl_products.id
LEFT JOIN tbl_products_categories ON tbl_products_categories.product_id = tbl_products.id
LEFT JOIN tbl_product_sales
	ON tbl_product_sales.product_id = tbl_products.id
	AND tbl_product_sales.is_active = 1
	AND datetime('now') BETWEEN tbl_product_sales.starts_at AND tbl_product_sales.ends_at
WHERE tbl_products_categories.category_id = ?
	AND tbl_products.id != ?
	AND tbl_products.status = 'ACTIVE'
	AND COALESCE(tbl_product_images.thumbnail, '') != ''
LIMIT 6;

-- name: GetProductsWithoutSlugs :many
SELECT
	tbl_products.id,
	tbl_products.name,
	tbl_products.serial,
	tbl_brands.name as brand_name,
	COALESCE(pc.category, '') AS product_category,
	COALESCE(pc.subcategory, '') AS product_subcategory,
	COALESCE(tbl_product_specs.power, '') AS power,
	COALESCE(tbl_product_specs.capacity, '') AS capacity
FROM tbl_products
INNER JOIN tbl_brands ON tbl_brands.id = tbl_products.brand_id
LEFT JOIN tbl_product_specs ON tbl_product_specs.id = tbl_products.product_specs_id
LEFT JOIN tbl_products_categories ON tbl_products_categories.product_id = tbl_products.id
LEFT JOIN tbl_product_categories AS pc ON pc.id = tbl_products_categories.category_id
WHERE
	slug = '' or slug IS NULL
	AND tbl_products.status = 'ACTIVE';

-- name: UpdateProductSlugByID :exec
UPDATE tbl_products SET slug = ? WHERE id = ?;
