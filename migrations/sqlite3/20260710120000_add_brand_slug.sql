-- +goose Up
-- +goose StatementBegin
ALTER TABLE tbl_brands ADD COLUMN slug TEXT;

CREATE INDEX IF NOT EXISTS idx_brands_slug ON tbl_brands(slug);

UPDATE tbl_brands
SET slug = lower(trim(name))
WHERE slug IS NULL OR slug = '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_brands_slug;
ALTER TABLE tbl_brands DROP COLUMN slug;
-- +goose StatementEnd
