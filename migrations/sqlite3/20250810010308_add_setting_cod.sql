-- +goose Up
-- +goose StatementBegin
INSERT INTO tbl_settings(name, value)
VALUES ('cash_on_delivery', 'false')
;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM tbl_settings
WHERE name = 'cash_on_delivery'
;
-- +goose StatementEnd
