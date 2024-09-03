-- name: GetBrandIDByName :one
SELECT id
FROM tbl_brand
WHERE
	name = ?
LIMIT 1;

-- name: GetBrandByID :one
SELECT
	tbl_brand.*,
	tbl_brand_image.id AS brand_image_id,
	tbl_brand_image.path AS path
FROM tbl_brand
INNER JOIN tbl_brand_image ON tbl_brand_image.brand_id = tbl_brand.id
WHERE
	tbl_brand.id = ?
LIMIT 1;

-- name: GetBrandLogos :many
SELECT
	tbl_brand.id AS id,
	tbl_brand.name AS name,
	tbl_brand_image.id AS brand_image_id,
	tbl_brand_image.path AS path
FROM tbl_brand
INNER JOIN tbl_brand_image ON tbl_brand_image.brand_id = tbl_brand.id
WHERE
	tbl_brand_image.is_main = true
ORDER BY tbl_brand.created_at DESC
LIMIT ?;

-- name: CreateBrand :one
INSERT INTO tbl_brand (
	name,
	created_at,
	updated_at,
	deleted_at
) VALUES (
	?, ?, ?, ?
) RETURNING id;

-- name: CreateBrandImage :one
INSERT INTO tbl_brand_image (
	brand_id,
	path,
	is_main,
	created_at,
	updated_at,
	deleted_at
) VALUES (
	?, ?, ?,
	?, ?, ?
) RETURNING id;
