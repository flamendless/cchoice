-- +goose Up
CREATE TABLE IF NOT EXISTS tbl_customer_otp_codes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    customer_id INTEGER NOT NULL,
    otp_code TEXT NOT NULL,
    expires_at TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    used_at TEXT,
    FOREIGN KEY (customer_id) REFERENCES tbl_customers(id),
    UNIQUE (customer_id, otp_code, used_at)
);

-- +goose Down
DROP TABLE IF EXISTS tbl_customer_otp_codes;
