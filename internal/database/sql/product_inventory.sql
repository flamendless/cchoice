-- name: CreateProductInventory :one
INSERT INTO tbl_product_inventories (
    product_id,
    stocks,
    stocks_in,
    created_at,
    updated_at
) VALUES (
    ?,
    ?,
    ?,
    DATETIME('now'),
    DATETIME('now')
) RETURNING id;

-- name: GetProductInventoryByProductID :one
SELECT
    id,
    product_id,
    stocks,
    stocks_in,
    created_at,
    updated_at
FROM tbl_product_inventories
WHERE product_id = ?
LIMIT 1;

-- name: GetProductInventoryByID :one
SELECT
    id,
    product_id,
    stocks,
    stocks_in,
    created_at,
    updated_at
FROM tbl_product_inventories
WHERE id = ?
LIMIT 1;

-- name: UpdateProductInventory :exec
UPDATE tbl_product_inventories
SET
    stocks = ?,
    updated_at = DATETIME('now')
WHERE product_id = ? AND stocks_in = ?;

-- name: ListProductInventories :many
SELECT
    tbl_product_inventories.id,
    tbl_product_inventories.product_id,
    tbl_product_inventories.stocks,
    tbl_product_inventories.stocks_in,
    tbl_product_inventories.created_at,
    tbl_product_inventories.updated_at,
    tbl_products.name AS product_name,
    tbl_products.slug AS product_slug
FROM tbl_product_inventories
INNER JOIN tbl_products ON tbl_products.id = tbl_product_inventories.product_id
ORDER BY tbl_products.name ASC
LIMIT ?;
