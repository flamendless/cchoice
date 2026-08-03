-- +goose Up
-- +goose StatementBegin
-- Scripts already applied on existing environments before tbl_datamigrate_applied tracking.
INSERT INTO tbl_datamigrate_applied (name, applied_at) VALUES
    ('apply_discount:sale_2025', datetime('now')),
    ('populate_product_images_cdn', datetime('now')),
    ('populate_product_slugs', datetime('now'));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM tbl_datamigrate_applied
WHERE name IN (
    'apply_discount:sale_2025',
    'populate_product_images_cdn',
    'populate_product_slugs'
);
-- +goose StatementEnd
