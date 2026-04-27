-- +goose Up
-- +goose StatementBegin
UPDATE tbl_settings SET value = 'https://cchoice.shop'
WHERE name = 'url_main_shop'
;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- +goose StatementEnd
