-- +goose Up
-- +goose StatementBegin
ALTER TABLE tbl_promos ADD COLUMN tracked_link_id TEXT NOT NULL DEFAULT '';
ALTER TABLE tbl_promos ADD COLUMN link_url TEXT NOT NULL DEFAULT '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE tbl_promos DROP COLUMN tracked_link_id;
ALTER TABLE tbl_promos DROP COLUMN link_url;
-- +goose StatementEnd
