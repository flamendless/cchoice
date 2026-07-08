-- name: GetProductExternalPlatformLinksByProductID :many
SELECT *
FROM tbl_product_external_platform_links
WHERE product_id = ?
ORDER BY platform ASC;

-- name: DeleteProductExternalPlatformLinksByProductID :exec
DELETE FROM tbl_product_external_platform_links
WHERE product_id = ?;

-- name: CreateProductExternalPlatformLink :one
INSERT INTO tbl_product_external_platform_links (
    product_id,
    platform,
    url,
    created_at,
    updated_at
) VALUES (
    ?,
    ?,
    ?,
    datetime('now'),
    datetime('now')
)
RETURNING *;
