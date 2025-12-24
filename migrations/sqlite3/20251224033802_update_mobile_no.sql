-- +goose Up
-- +goose StatementBegin
UPDATE tbl_settings SET value = '+639976894824' WHERE name = 'mobile_no';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- +goose StatementEnd
