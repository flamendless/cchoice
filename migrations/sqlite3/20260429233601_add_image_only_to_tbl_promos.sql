-- +goose Up
-- +goose StatementBegin
ALTER TABLE tbl_promos ADD COLUMN banner_only BOOLEAN DEFAULT TRUE;
ALTER TABLE tbl_promos ADD COLUMN priority INTEGER DEFAULT 1;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE tbl_promos DROP COLUMN banner_only;
ALTER TABLE tbl_promos DROP COLUMN priority;
-- +goose StatementEnd
