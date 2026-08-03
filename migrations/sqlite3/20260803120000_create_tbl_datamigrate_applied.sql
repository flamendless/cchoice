-- +goose Up
-- +goose StatementBegin
CREATE TABLE tbl_datamigrate_applied (
    name TEXT PRIMARY KEY,
    applied_at DATETIME NOT NULL DEFAULT (datetime('now'))
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE tbl_datamigrate_applied;
-- +goose StatementEnd
