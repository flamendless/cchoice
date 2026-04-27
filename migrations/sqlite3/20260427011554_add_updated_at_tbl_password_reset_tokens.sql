-- +goose Up
-- +goose StatementBegin
ALTER TABLE tbl_password_reset_tokens ADD COLUMN updated_at DATETIME;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE tbl_password_reset_tokens DROP COLUMN updated_at;
-- +goose StatementEnd
