-- +goose Up
-- +goose StatementBegin
ALTER TABLE tbl_brands ADD COLUMN slug TEXT;

CREATE INDEX IF NOT EXISTS idx_brands_slug ON tbl_brands(slug);

-- Slug backfill: trim, lowercase, replace whitespace with hyphens
UPDATE tbl_brands
SET slug = lower(trim(name))
WHERE slug IS NULL OR slug = '';

UPDATE tbl_brands
SET slug = replace(slug, ' ', '-')
WHERE slug IS NOT NULL;

UPDATE tbl_brands
SET slug = replace(slug, char(9), '-')
WHERE slug IS NOT NULL;

UPDATE tbl_brands
SET slug = replace(slug, char(10), '-')
WHERE slug IS NOT NULL;

UPDATE tbl_brands
SET slug = replace(slug, char(13), '-')
WHERE slug IS NOT NULL;

-- Collapse repeated hyphens from multiple spaces/tabs
UPDATE tbl_brands SET slug = replace(slug, '--', '-') WHERE slug LIKE '%--%';
UPDATE tbl_brands SET slug = replace(slug, '--', '-') WHERE slug LIKE '%--%';
UPDATE tbl_brands SET slug = replace(slug, '--', '-') WHERE slug LIKE '%--%';
UPDATE tbl_brands SET slug = replace(slug, '--', '-') WHERE slug LIKE '%--%';

UPDATE tbl_brands
SET slug = trim(slug, '-')
WHERE slug IS NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_brands_slug;
ALTER TABLE tbl_brands DROP COLUMN slug;
-- +goose StatementEnd
