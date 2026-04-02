-- name: CreateCpoint :one
INSERT INTO tbl_cpoints (
    customer_id,
    code,
    value,
    product_skus,
    expires_at,
    generated_at,
    created_at,
    updated_at,
    deleted_at
) VALUES (
    ?, ?, ?, ?, ?, datetime('now'), datetime('now'), datetime('now'), '1970-01-01 00:00:00+00:00'
) RETURNING id;

-- name: GetCpointsByCustomerID :many
SELECT
    id,
    customer_id,
    code,
    value,
    product_skus,
    expires_at,
    generated_at,
    redeemed_at,
    created_at,
    updated_at
FROM tbl_cpoints
WHERE
    customer_id = ?
    AND deleted_at = '1970-01-01 00:00:00+00:00'
ORDER BY created_at DESC;

-- name: GetCpointsByCustomerIDWithTotal :many
SELECT
    c.id,
    c.customer_id,
    c.code,
    c.value,
    c.product_skus,
    c.expires_at,
    c.generated_at,
    c.redeemed_at,
    c.created_at,
    c.updated_at,
    COUNT(*) OVER() AS total
FROM tbl_cpoints c
WHERE
    c.customer_id = ?
    AND c.deleted_at = '1970-01-01 00:00:00+00:00'
ORDER BY c.created_at DESC;

-- name: RedeemCpoint :one
UPDATE tbl_cpoints
SET
    redeemed_at = datetime('now'),
    updated_at = datetime('now')
WHERE
    code = ?
    AND redeemed_at IS NULL
    AND deleted_at = '1970-01-01 00:00:00+00:00'
RETURNING id, customer_id, code, value, product_skus, expires_at, generated_at, redeemed_at;

-- name: GetCpointByCode :one
SELECT
    id,
    customer_id,
    code,
    value,
    product_skus,
    expires_at,
    generated_at,
    redeemed_at,
    created_at,
    updated_at
FROM tbl_cpoints
WHERE
    code = ?
    AND deleted_at = '1970-01-01 00:00:00+00:00'
LIMIT 1;
