-- +goose Up
ALTER TABLE tbl_staffs ADD COLUMN require_in_shop BOOLEAN NOT NULL DEFAULT 0;

-- +goose Down
ALTER TABLE tbl_staffs DROP COLUMN require_in_shop;
