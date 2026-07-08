-- +goose Up
-- +goose StatementBegin
CREATE TABLE tbl_product_external_platform_links (
    id INTEGER PRIMARY KEY,
    product_id INTEGER NOT NULL,
    platform TEXT NOT NULL,
    url TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT (datetime('now')),
    updated_at DATETIME NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (product_id) REFERENCES tbl_products(id),
    UNIQUE (product_id, platform)
);

CREATE INDEX idx_product_external_platform_links_product_id ON tbl_product_external_platform_links(product_id);
-- +goose StatementEnd

-- +goose Down
DROP INDEX IF EXISTS idx_product_external_platform_links_product_id;
DROP TABLE IF EXISTS tbl_product_external_platform_links;
