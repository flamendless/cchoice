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

-- name: AdminGetProductInventoriesListing :many
SELECT
    tbl_product_inventories.id,
    tbl_product_inventories.stocks,
    tbl_product_inventories.stocks_in,
    tbl_product_inventories.updated_at,
    tbl_products.serial AS product_serial,
    tbl_products.slug AS product_slug,
    tbl_products.status AS product_status,
    tbl_products.name AS product_name,
    tbl_brands.name AS brand_name
FROM tbl_product_inventories
INNER JOIN tbl_products ON tbl_products.id = tbl_product_inventories.product_id
INNER JOIN tbl_brands ON tbl_brands.id = tbl_products.brand_id
WHERE
    (@search_serial IS NULL OR @search_serial = '' OR LOWER(tbl_products.serial) LIKE '%' || LOWER(@search_serial) || '%')
    AND (@search_brand IS NULL OR @search_brand = '' OR LOWER(tbl_brands.name) LIKE '%' || LOWER(@search_brand) || '%')
    AND (@product_status IS NULL OR @product_status = '' OR tbl_products.status = @product_status)
    AND (@stocks_in IS NULL OR @stocks_in = '' OR tbl_product_inventories.stocks_in = @stocks_in)
ORDER BY
    CASE
        WHEN tbl_product_inventories.stocks = 0 THEN 0
        WHEN tbl_product_inventories.stocks <= 10 THEN 1
        ELSE 2
    END ASC,
    tbl_product_inventories.stocks ASC,
    tbl_product_inventories.updated_at DESC;
