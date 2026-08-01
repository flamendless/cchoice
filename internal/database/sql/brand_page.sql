-- name: GetBrandsForListingPage :many
SELECT
	tbl_brands.id AS id,
	tbl_brands.name AS name,
	tbl_brands.slug AS slug,
	tbl_brand_images.path AS path,
	tbl_brand_images.s3_url AS s3_url,
	COUNT(tbl_products.id) AS product_count
FROM tbl_brands
LEFT JOIN tbl_brand_images ON tbl_brand_images.brand_id = tbl_brands.id AND tbl_brand_images.is_main = true
LEFT JOIN tbl_products ON tbl_products.brand_id = tbl_brands.id
	AND tbl_products.status = 'ACTIVE'
	AND tbl_products.deleted_at = '1970-01-01 00:00:00+00:00'
WHERE
	tbl_brands.status = 'ACTIVE'
	AND tbl_brands.deleted_at = '1970-01-01 00:00:00+00:00'
GROUP BY tbl_brands.id, tbl_brand_images.id
ORDER BY product_count DESC, tbl_brands.name ASC;

-- name: GetBrandBySlug :one
SELECT
	tbl_brands.id,
	tbl_brands.name,
	tbl_brands.slug,
	tbl_brands.status,
	tbl_brands.created_at,
	tbl_brands.updated_at,
	tbl_brands.deleted_at,
	tbl_brand_images.id AS brand_image_id,
	tbl_brand_images.path AS path,
	tbl_brand_images.s3_url AS s3_url
FROM tbl_brands
LEFT JOIN tbl_brand_images ON tbl_brand_images.brand_id = tbl_brands.id AND tbl_brand_images.is_main = true
WHERE
	tbl_brands.slug = ?
	AND tbl_brands.status = 'ACTIVE'
	AND tbl_brands.deleted_at = '1970-01-01 00:00:00+00:00'
LIMIT 1;

-- name: BrandSlugExists :one
SELECT COUNT(*) > 0 AS brand_exists
FROM tbl_brands
WHERE
	slug = ?
	AND status = 'ACTIVE'
	AND deleted_at = '1970-01-01 00:00:00+00:00';

-- name: GetBestSellingProductsByBrandID :many
SELECT
	tbl_products.id,
	tbl_products.serial,
	tbl_products.slug,
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
	tbl_product_images.cdn_url_thumbnail,
	COALESCE(sales.total_quantity, 0) AS total_sold
FROM tbl_products
INNER JOIN tbl_brands ON tbl_brands.id = tbl_products.brand_id
LEFT JOIN tbl_product_images ON tbl_product_images.product_id = tbl_products.id
LEFT JOIN tbl_product_sales
	ON tbl_product_sales.product_id = tbl_products.id
	AND tbl_product_sales.is_active = 1
	AND datetime('now') BETWEEN
		tbl_product_sales.starts_at AND tbl_product_sales.ends_at
LEFT JOIN (
	SELECT
		tbl_order_lines.product_id,
		SUM(tbl_order_lines.quantity) AS total_quantity
	FROM tbl_order_lines
	INNER JOIN tbl_orders ON tbl_orders.id = tbl_order_lines.order_id
	WHERE tbl_orders.status NOT IN ('CANCELLED', 'REFUNDED')
	GROUP BY tbl_order_lines.product_id
) AS sales ON sales.product_id = tbl_products.id
WHERE
	tbl_products.status = 'ACTIVE'
	AND tbl_products.brand_id = ?
	AND COALESCE(tbl_product_images.thumbnail, 'static/images/empty_96x96.webp') != 'static/images/empty_96x96.webp'
ORDER BY total_sold DESC, tbl_products.created_at DESC
LIMIT ?;

-- name: GetHighestDiscountProductsByBrandID :many
SELECT
	tbl_products.id,
	tbl_products.serial,
	tbl_products.slug,
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
	tbl_product_images.cdn_url_thumbnail,
	CASE
		WHEN tbl_product_sales.id IS NOT NULL AND tbl_products.unit_price_with_vat > 0
			THEN CAST(
				((tbl_products.unit_price_with_vat - tbl_product_sales.sale_price_with_vat) * 100.0)
				/ tbl_products.unit_price_with_vat AS INTEGER
			)
		ELSE 0
	END AS discount_percent
FROM tbl_products
INNER JOIN tbl_brands ON tbl_brands.id = tbl_products.brand_id
LEFT JOIN tbl_product_images ON tbl_product_images.product_id = tbl_products.id
INNER JOIN tbl_product_sales
	ON tbl_product_sales.product_id = tbl_products.id
	AND tbl_product_sales.is_active = 1
	AND datetime('now') BETWEEN
		tbl_product_sales.starts_at AND tbl_product_sales.ends_at
WHERE
	tbl_products.status = 'ACTIVE'
	AND tbl_products.brand_id = ?
	AND COALESCE(tbl_product_images.thumbnail, 'static/images/empty_96x96.webp') != 'static/images/empty_96x96.webp'
ORDER BY discount_percent DESC, tbl_products.created_at DESC
LIMIT ?;

-- name: GetBrandProductCategoriesForSections :many
SELECT
	tbl_product_categories.id,
	tbl_product_categories.category,
	tbl_product_categories.subcategory,
	COUNT(tbl_products_categories.product_id) AS products_count
FROM tbl_product_categories
INNER JOIN tbl_products_categories ON tbl_products_categories.category_id = tbl_product_categories.id
INNER JOIN tbl_products ON tbl_products.id = tbl_products_categories.product_id
LEFT JOIN tbl_product_images ON tbl_product_images.product_id = tbl_products.id
WHERE
	tbl_products.brand_id = ?
	AND tbl_products.status = 'ACTIVE'
	AND tbl_products.deleted_at = '1970-01-01 00:00:00+00:00'
	AND tbl_product_categories.category IS NOT NULL
	AND tbl_product_categories.category != ''
	AND tbl_product_categories.subcategory IS NOT NULL
	AND tbl_product_categories.subcategory != ''
	AND COALESCE(tbl_product_images.thumbnail, 'static/images/empty_96x96.webp') != 'static/images/empty_96x96.webp'
GROUP BY tbl_product_categories.id
HAVING products_count > 0
ORDER BY tbl_product_categories.category ASC, tbl_product_categories.subcategory ASC;

-- name: GetProductsByBrandAndCategoryID :many
SELECT
	tbl_products.id,
	tbl_products.serial,
	tbl_products.slug,
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
INNER JOIN tbl_products_categories ON tbl_products_categories.product_id = tbl_products.id
LEFT JOIN tbl_product_images ON tbl_product_images.product_id = tbl_products.id
LEFT JOIN tbl_product_sales
	ON tbl_product_sales.product_id = tbl_products.id
	AND tbl_product_sales.is_active = 1
	AND datetime('now') BETWEEN
		tbl_product_sales.starts_at AND tbl_product_sales.ends_at
WHERE
	tbl_products.status = 'ACTIVE'
	AND tbl_products.brand_id = ?
	AND tbl_products_categories.category_id = ?
	AND COALESCE(tbl_product_images.thumbnail, 'static/images/empty_96x96.webp') != 'static/images/empty_96x96.webp'
ORDER BY is_on_sale DESC, tbl_products.created_at DESC
LIMIT ?;

-- name: ListBrandSitemapSlugs :many
SELECT DISTINCT tbl_brands.slug
FROM tbl_brands
INNER JOIN tbl_products ON tbl_products.brand_id = tbl_brands.id
LEFT JOIN tbl_product_images ON tbl_product_images.product_id = tbl_products.id
WHERE
	tbl_brands.status = 'ACTIVE'
	AND tbl_brands.deleted_at = '1970-01-01 00:00:00+00:00'
	AND tbl_brands.slug IS NOT NULL
	AND tbl_brands.slug != ''
	AND tbl_products.status = 'ACTIVE'
	AND tbl_products.deleted_at = '1970-01-01 00:00:00+00:00'
	AND COALESCE(tbl_product_images.thumbnail, 'static/images/empty_96x96.webp') != 'static/images/empty_96x96.webp'
ORDER BY tbl_brands.name ASC;
