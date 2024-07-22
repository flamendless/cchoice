-- name: GetBrandIDByName :one
SELECT id
FROM tbl_brand
WHERE
	name = ?
LIMIT 1;

-- name: CreateBrand :one
INSERT INTO tbl_brand (
	name
) VALUES (
	?
) RETURNING *;
