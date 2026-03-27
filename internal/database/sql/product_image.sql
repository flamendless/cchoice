-- name: CreateProductImage :one
INSERT INTO tbl_product_images (
	product_id,
	path,
	thumbnail,
	cdn_url,
	cdn_url_thumbnail,
	created_at,
	updated_at,
	deleted_at
) VALUES (
	?, ?, ?, ?, ?, ?, ?, ?
) RETURNING *;

-- name: GetProductImageByProductID :one
SELECT * FROM tbl_product_images
WHERE product_id = ?
LIMIT 1;

-- name: UpdateProductImageThumbnail :one
UPDATE tbl_product_images
SET thumbnail = ?, updated_at = datetime('now')
WHERE id = ?
RETURNING *;

-- name: UpdateProductImageCDNURLs :one
UPDATE tbl_product_images
SET cdn_url = ?, cdn_url_thumbnail = ?, updated_at = datetime('now')
WHERE id = ?
RETURNING *;

-- name: GetAllProductImages :many
SELECT * FROM tbl_product_images;

-- name: GetProductImagesWithEmptyCDNURLs :many
SELECT
	id,
	product_id,
	COALESCE(
		thumbnail,
		'static/images/empty_96x96.webp'
	) AS thumbnail_path,
	cdn_url,
	cdn_url_thumbnail
FROM tbl_product_images
WHERE cdn_url = '' OR cdn_url_thumbnail = '';

-- name: GetProductImagesWithEmptyCDNURLsForce :many
SELECT
	id,
	product_id,
	COALESCE(
		thumbnail,
		'static/images/empty_96x96.webp'
	) AS thumbnail_path,
	cdn_url,
	cdn_url_thumbnail
FROM tbl_product_images;
