-- +goose Up
-- +goose StatementBegin
ALTER TABLE tbl_checkouts ADD COLUMN status TEXT NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'COMPLETED', 'CANCELLED'));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE tbl_checkouts DROP COLUMN status;
-- +goose StatementEnd

