-- +goose Up
-- +goose StatementBegin
ALTER TABLE tbl_brands ADD COLUMN status TEXT NOT NULL DEFAULT 'DRAFT';

UPDATE tbl_brands SET status = 'ACTIVE' WHERE deleted_at = '1970-01-01 00:00:00+00:00';
UPDATE tbl_brands SET status = 'DELETED' WHERE deleted_at != '1970-01-01 00:00:00+00:00';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE tbl_brands DROP COLUMN status;
-- +goose StatementEnd
