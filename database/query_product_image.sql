-- name: CreateProductImage :one
INSERT INTO tbl_product_image (
	product_id,
	path,
	created_at,
	updated_at,
	deleted_at
) VALUES (
	?, ?, ?, ?, ?
) RETURNING *;
