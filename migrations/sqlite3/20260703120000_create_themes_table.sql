-- +goose Up
CREATE TABLE IF NOT EXISTS tbl_themes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'DRAFT',
    start_date TEXT NOT NULL,
    end_date TEXT NOT NULL,
    configuration TEXT NOT NULL,
    configuration_type TEXT NOT NULL DEFAULT 'JSON',
    created_by INTEGER NOT NULL REFERENCES tbl_staffs(id),
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now')),
    deleted_at TEXT NOT NULL DEFAULT '1970-01-01 00:00:00+00:00'
);

CREATE INDEX IF NOT EXISTS idx_themes_status ON tbl_themes(status);
CREATE INDEX IF NOT EXISTS idx_themes_deleted_at ON tbl_themes(deleted_at);
CREATE INDEX IF NOT EXISTS idx_themes_start_date ON tbl_themes(start_date);
CREATE INDEX IF NOT EXISTS idx_themes_end_date ON tbl_themes(end_date);
CREATE INDEX IF NOT EXISTS idx_themes_created_by ON tbl_themes(created_by);

-- +goose Down
DROP TABLE IF EXISTS tbl_themes;
