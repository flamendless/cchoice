-- +goose Up
ALTER TABLE tbl_memos ADD COLUMN file_url TEXT;

-- +goose Down
