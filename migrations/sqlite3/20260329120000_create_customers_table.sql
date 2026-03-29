-- +goose Up
CREATE TABLE IF NOT EXISTS tbl_customers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    first_name TEXT NOT NULL,
    middle_name TEXT,
    last_name TEXT NOT NULL,
    birthdate TEXT NOT NULL,
    sex TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    mobile_no TEXT NOT NULL,
    password TEXT NOT NULL,
    customer_type TEXT NOT NULL CHECK(customer_type IN ('customer', 'company')),
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now')),
    deleted_at TEXT NOT NULL DEFAULT ('1970-01-01 00:00:00+00:00')
);

CREATE TABLE IF NOT EXISTS tbl_customer_companies (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    customer_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now')),
    deleted_at TEXT NOT NULL DEFAULT ('1970-01-01 00:00:00+00:00'),
    FOREIGN KEY (customer_id) REFERENCES tbl_customers(id)
);

-- +goose Down
DROP TABLE IF EXISTS tbl_customer_companies;
DROP TABLE IF EXISTS tbl_customers;
