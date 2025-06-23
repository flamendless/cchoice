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
