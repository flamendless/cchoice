-- +goose Up
ALTER TABLE tbl_staffs ADD COLUMN status TEXT NOT NULL DEFAULT 'PROBATION';

-- +goose Down
ALTER TABLE tbl_staffs DROP COLUMN status;
