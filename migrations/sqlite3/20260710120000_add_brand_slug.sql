-- +goose Up
-- +goose StatementBegin
ALTER TABLE tbl_brands ADD COLUMN slug TEXT;

CREATE INDEX IF NOT EXISTS idx_brands_slug ON tbl_brands(slug);
-- +goose StatementEnd

-- Populate slugs with: go run ./main.go populate_brand_slugs --dry-run=false

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_brands_slug;
ALTER TABLE tbl_brands DROP COLUMN slug;
-- +goose StatementEnd
