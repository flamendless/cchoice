-- +goose Up
-- +goose StatementBegin
CREATE TABLE tbl_quotation_lines (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    quotation_id INTEGER NOT NULL,
    product_id INTEGER NOT NULL,
    quantity INTEGER NOT NULL DEFAULT 1,
    original_price_snapshot INTEGER DEFAULT 0,
    sale_price_snapshot INTEGER DEFAULT 0,
    currency TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT (datetime('now')),
    updated_at DATETIME NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (quotation_id) REFERENCES tbl_quotations(id) ON DELETE CASCADE,
    FOREIGN KEY (product_id) REFERENCES tbl_products(id)
);

CREATE TABLE tbl_quotations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    customer_id INTEGER NOT NULL,
    acknowledged_by_staff_id INTEGER,
    status TEXT NOT NULL DEFAULT 'DRAFT',
    created_at DATETIME NOT NULL DEFAULT (datetime('now')),
    updated_at DATETIME NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (customer_id) REFERENCES tbl_customers(id),
    FOREIGN KEY (acknowledged_by_staff_id) REFERENCES tbl_staffs(id)
);

CREATE INDEX idx_quotation_lines ON tbl_quotation_lines(quotation_id);
CREATE INDEX idx_quotations_customer_status ON tbl_quotations(customer_id, status);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_quotation_lines;
DROP INDEX IF EXISTS idx_quotations_customer_status;
DROP TABLE IF EXISTS tbl_quotation_lines;
DROP TABLE IF EXISTS tbl_quotations;
-- +goose StatementEnd
