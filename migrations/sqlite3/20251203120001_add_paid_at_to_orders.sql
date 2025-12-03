-- +goose Up
-- +goose StatementBegin
ALTER TABLE tbl_orders ADD COLUMN paid_at DATETIME;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE tbl_orders DROP COLUMN paid_at;
-- +goose StatementEnd

