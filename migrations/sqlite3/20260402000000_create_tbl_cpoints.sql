-- +goose Up
CREATE TABLE IF NOT EXISTS tbl_cpoints (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    customer_id INTEGER NOT NULL,
    code TEXT NOT NULL UNIQUE,
    value INTEGER NOT NULL,
    product_skus TEXT,
    expires_at TEXT,
    generated_at TEXT NOT NULL DEFAULT (datetime('now')),
    redeemed_at TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now')),
    deleted_at TEXT NOT NULL DEFAULT ('1970-01-01 00:00:00+00:00'),
    FOREIGN KEY (customer_id) REFERENCES tbl_customers(id)
);

CREATE INDEX IF NOT EXISTS idx_cpoints_customer_id ON tbl_cpoints(customer_id);

-- +goose Down
DROP INDEX IF EXISTS idx_cpoints_customer_id;
DROP TABLE IF EXISTS tbl_cpoints;
