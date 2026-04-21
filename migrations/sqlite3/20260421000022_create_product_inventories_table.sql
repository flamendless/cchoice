-- +goose Up
-- +goose StatementBegin
CREATE TABLE tbl_product_inventories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    product_id INTEGER NOT NULL UNIQUE,
    stocks INTEGER NOT NULL DEFAULT 0,
    stocks_in TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (product_id) REFERENCES tbl_products(id)
);

CREATE INDEX idx_product_inventories_product_id ON tbl_product_inventories(product_id);
-- +goose StatementEnd

-- +goose Down
DROP INDEX IF EXISTS idx_product_inventories_product_id;
DROP TABLE IF EXISTS tbl_product_inventories;
