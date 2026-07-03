-- +goose Up
-- +goose StatementBegin
INSERT INTO tbl_settings(name, value)
VALUES
	('url_lazada', 'https://www.lazada.com.ph/shop/cme-apparel')
;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM tbl_settings
WHERE name IN ('url_lazada');
-- +goose StatementEnd
