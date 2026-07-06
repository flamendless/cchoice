-- +goose Up
ALTER TABLE tbl_orders ADD COLUMN customer_id INTEGER REFERENCES tbl_customers(id);
ALTER TABLE tbl_orders ADD COLUMN shipping_barangay TEXT NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_orders_customer_id ON tbl_orders(customer_id);

-- +goose Down
DROP INDEX IF EXISTS idx_orders_customer_id;
ALTER TABLE tbl_orders DROP COLUMN shipping_barangay;
ALTER TABLE tbl_orders DROP COLUMN customer_id;
