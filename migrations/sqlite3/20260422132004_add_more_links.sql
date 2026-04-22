-- +goose Up
-- +goose StatementBegin
INSERT INTO tbl_settings(name, value)
VALUES
	('url_main_shop', 'https://cchoice.shop/l/platforms')
;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM tbl_settings
WHERE name IN (
	'url_main_shop'
);
-- +goose StatementEnd
