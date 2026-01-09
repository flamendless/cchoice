-- +goose Up
-- +goose StatementBegin
CREATE TABLE tbl_product_sales (
	id INTEGER PRIMARY KEY,
	product_id INTEGER NOT NULL,

	sale_price_without_vat INTEGER NOT NULL,
	sale_price_with_vat INTEGER NOT NULL,

	sale_price_without_vat_currency TEXT NOT NULL,
	sale_price_with_vat_currency TEXT NOT NULL,

	discount_type TEXT NOT NULL CHECK (discount_type IN ('fixed', 'percentage')),
	discount_value INTEGER NOT NULL, -- cents if fixed, percent if percentage

	starts_at DATETIME NOT NULL,
	ends_at DATETIME NOT NULL,

	is_active BOOLEAN NOT NULL DEFAULT 0,

	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	deleted_at DATETIME NOT NULL DEFAULT (DATE('1970-01-01 00:00:00')),

	FOREIGN KEY (product_id) REFERENCES tbl_products(id) ON DELETE CASCADE
);

CREATE INDEX idx_product_sales_active
ON tbl_product_sales (product_id, is_active, starts_at, ends_at);

CREATE INDEX idx_product_sales_active_window
ON tbl_product_sales (product_id, is_active, starts_at, ends_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX idx_product_sales_active;
DROP INDEX idx_product_sales_active_window;
DROP TABLE tbl_product_sales;
-- +goose StatementEnd
