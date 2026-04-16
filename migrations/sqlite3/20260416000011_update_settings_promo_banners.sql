-- +goose Up
-- +goose StatementBegin
UPDATE tbl_settings
SET name = 'show_random_sale_product'
WHERE name = 'show_promo_banner';

INSERT INTO tbl_settings(name, value)
VALUES('show_promo_banners', 'true');

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM tbl_settings WHERE name = 'show_promo_banners';

UPDATE tbl_settings
SET name = 'show_promo_banner'
WHERE name = 'show_random_sale_product';
-- +goose StatementEnd
