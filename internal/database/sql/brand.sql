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
