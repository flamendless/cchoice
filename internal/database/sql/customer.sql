-- name: GetCustomerByEmail :one
SELECT
    id,
    first_name,
    middle_name,
    last_name,
    birthdate,
    sex,
    email,
    mobile_no,
    password,
    customer_type,
    created_at,
    updated_at
FROM tbl_customers
WHERE
    email = ?
    AND deleted_at = '1970-01-01 00:00:00+00:00'
LIMIT 1;

-- name: GetCustomerByID :one
SELECT
    id,
    first_name,
    middle_name,
    last_name,
    birthdate,
    sex,
    email,
    mobile_no,
    password,
    customer_type,
    created_at,
    updated_at
FROM tbl_customers
WHERE
    id = ?
    AND deleted_at = '1970-01-01 00:00:00+00:00'
LIMIT 1;

-- name: GetCustomerCompanyByCustomerID :one
SELECT
    id,
    customer_id,
    name,
    created_at,
    updated_at
FROM tbl_customer_companies
WHERE
    customer_id = ?
    AND deleted_at = '1970-01-01 00:00:00+00:00'
LIMIT 1;

-- name: CreateCustomer :one
INSERT INTO tbl_customers (
    first_name,
    middle_name,
    last_name,
    birthdate,
    sex,
    email,
    mobile_no,
    password,
    customer_type,
    created_at,
    updated_at,
    deleted_at
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now'), datetime('now'), '1970-01-01 00:00:00+00:00'
) RETURNING id;

-- name: CreateCustomerCompany :one
INSERT INTO tbl_customer_companies (
    customer_id,
    name,
    created_at,
    updated_at,
    deleted_at
) VALUES (
    ?, ?, datetime('now'), datetime('now'), '1970-01-01 00:00:00+00:00'
) RETURNING id;

-- name: UpdateCustomerPassword :one
UPDATE tbl_customers
SET
    password = ?,
    updated_at = datetime('now')
WHERE
    id = ?
RETURNING id;

-- name: UpdateCustomerProfile :one
UPDATE tbl_customers
SET
    first_name = ?,
    middle_name = ?,
    last_name = ?,
    mobile_no = ?,
    birthdate = ?,
    sex = ?,
    updated_at = datetime('now')
WHERE
    id = ?
RETURNING id;
