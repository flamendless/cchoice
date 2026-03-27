-- name: CreateProductImage :one
INSERT INTO tbl_product_images (
	product_id,
	path,
	thumbnail,
	created_at,
	updated_at,
	deleted_at
) VALUES (
	?, ?, ?, ?, ?, ?
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
