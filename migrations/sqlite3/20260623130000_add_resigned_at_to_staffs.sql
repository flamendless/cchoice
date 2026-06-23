-- +goose Up
ALTER TABLE tbl_staffs ADD COLUMN resigned_at TEXT;

-- +goose Down
ALTER TABLE tbl_staffs DROP COLUMN resigned_at;
