-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS tbl_staff_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    staff_id INTEGER NOT NULL REFERENCES tbl_staffs(id),
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    action TEXT NOT NULL,
    module TEXT NOT NULL,
    result TEXT NOT NULL,
    useragent_id INTEGER REFERENCES tbl_useragents(id)
);
CREATE INDEX IF NOT EXISTS idx_staff_logs_staff_id ON tbl_staff_logs(staff_id);
CREATE INDEX IF NOT EXISTS idx_staff_logs_created_at ON tbl_staff_logs(created_at);
CREATE INDEX IF NOT EXISTS idx_staff_logs_module ON tbl_staff_logs(module);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_staff_logs_module;
DROP INDEX IF EXISTS idx_staff_logs_created_at;
DROP INDEX IF EXISTS idx_staff_logs_staff_id;
DROP TABLE IF EXISTS tbl_staff_logs;
-- +goose StatementEnd
