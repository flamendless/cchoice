-- +goose Up
-- +goose StatementBegin
ALTER TABLE tbl_products ADD COLUMN slug TEXT;

CREATE INDEX IF NOT EXISTS idx_products_slug ON tbl_products(slug);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_products_slug;
ALTER TABLE tbl_products DROP COLUMN slug;
-- +goose StatementEnd
