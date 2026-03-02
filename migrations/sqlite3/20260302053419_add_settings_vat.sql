-- +goose Up
-- +goose StatementBegin
INSERT INTO tbl_settings(name, value)
VALUES ('vat_percentage', '12');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM tbl_settings WHERE name = 'vat_percentage';
-- +goose StatementEnd
