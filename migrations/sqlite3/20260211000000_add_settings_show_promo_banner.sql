-- +goose Up
-- +goose StatementBegin
INSERT INTO tbl_settings(name, value)
VALUES
	('show_promo_banner', 'true')
;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM tbl_settings
WHERE name IN ('show_promo_banner');
-- +goose StatementEnd
