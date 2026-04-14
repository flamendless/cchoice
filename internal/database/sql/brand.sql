-- name: GetBrandsIDByName :one
SELECT id
FROM tbl_brands
WHERE
	name = ?
LIMIT 1;

-- name: GetBrandImageS3URLByPath :one
SELECT s3_url
FROM tbl_brand_images
WHERE
	path = ?
LIMIT 1;

-- name: GetBrandsByID :one
SELECT
	tbl_brands.*,
	tbl_brand_images.id AS brand_image_id,
	tbl_brand_images.path AS path,
	tbl_brand_images.s3_url AS s3_url
FROM tbl_brands
INNER JOIN tbl_brand_images ON tbl_brand_images.brand_id = tbl_brands.id
WHERE
	tbl_brands.id = ?
LIMIT 1;

-- name: GetBrandsLogos :many
SELECT
	tbl_brands.id AS id,
	tbl_brands.name AS name,
	tbl_brand_images.id AS brand_image_id,
	tbl_brand_images.path AS path,
	tbl_brand_images.s3_url AS s3_url
FROM tbl_brands
INNER JOIN tbl_brand_images ON tbl_brand_images.brand_id = tbl_brands.id
WHERE
	tbl_brand_images.is_main = true
ORDER BY tbl_brands.created_at DESC
LIMIT ?;

-- name: CreateBrands :one
INSERT INTO tbl_brands (
	name,
	created_at,
	updated_at,
	deleted_at
) VALUES (
	?, ?, ?, ?
) RETURNING id;

-- name: CreateBrandImages :one
INSERT INTO tbl_brand_images (
	brand_id,
	path,
	s3_url,
	is_main,
	created_at,
	updated_at,
	deleted_at
) VALUES (
	?, ?, ?, ?,
	?, ?, ?
) RETURNING id;

-- name: GetBrandsForSidePanel :many
SELECT
	id,
	name
FROM tbl_brands
WHERE deleted_at = '1970-01-01 00:00:00+00:00'
ORDER BY name ASC
LIMIT ?;

-- name: GetAllBrands :many
SELECT
	tbl_brands.id AS id,
	tbl_brands.name AS name,
	tbl_brand_images.id AS brand_image_id,
	tbl_brand_images.path AS path,
	tbl_brand_images.s3_url AS s3_url,
	tbl_brands.created_at AS created_at,
	COUNT(tbl_products.id) AS product_count
FROM tbl_brands
LEFT JOIN tbl_brand_images ON tbl_brand_images.brand_id = tbl_brands.id AND tbl_brand_images.is_main = true
LEFT JOIN tbl_products ON tbl_products.brand_id = tbl_brands.id AND tbl_products.deleted_at = '1970-01-01 00:00:00+00:00'
WHERE tbl_brands.deleted_at = '1970-01-01 00:00:00+00:00'
GROUP BY tbl_brands.id, tbl_brand_images.id
ORDER BY tbl_brands.name ASC;

-- name: SearchBrandsByName :many
SELECT
	tbl_brands.id AS id,
	tbl_brands.name AS name,
	tbl_brand_images.id AS brand_image_id,
	tbl_brand_images.path AS path,
	tbl_brand_images.s3_url AS s3_url,
	tbl_brands.created_at AS created_at,
	COUNT(tbl_products.id) AS product_count
FROM tbl_brands
LEFT JOIN tbl_brand_images ON tbl_brand_images.brand_id = tbl_brands.id AND tbl_brand_images.is_main = true
LEFT JOIN tbl_products ON tbl_products.brand_id = tbl_brands.id AND tbl_products.deleted_at = '1970-01-01 00:00:00+00:00'
WHERE
	tbl_brands.deleted_at = '1970-01-01 00:00:00+00:00'
	AND LOWER(tbl_brands.name) LIKE LOWER(?)
GROUP BY tbl_brands.id, tbl_brand_images.id
ORDER BY tbl_brands.name ASC;

-- name: UpdateBrand :exec
UPDATE tbl_brands
SET
	name = ?,
	updated_at = ?
WHERE
	id = ?
	AND deleted_at = '1970-01-01 00:00:00+00:00';

-- name: SoftDeleteBrand :exec
UPDATE tbl_brands
SET
	deleted_at = ?
WHERE
	id = ?
	AND deleted_at = '1970-01-01 00:00:00+00:00';
