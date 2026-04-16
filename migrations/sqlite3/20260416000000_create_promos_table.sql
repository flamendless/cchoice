-- +goose Up
CREATE TABLE IF NOT EXISTS tbl_promos (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    media_url TEXT NOT NULL,
    start_date TEXT NOT NULL,
    end_date TEXT NOT NULL,
    type TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'DRAFT',
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now')),
    deleted_at TEXT NOT NULL DEFAULT '1970-01-01 00:00:00+00:00'
);

CREATE INDEX IF NOT EXISTS idx_promos_status ON tbl_promos(status);
CREATE INDEX IF NOT EXISTS idx_promos_deleted_at ON tbl_promos(deleted_at);
CREATE INDEX IF NOT EXISTS idx_promos_start_date ON tbl_promos(start_date);
CREATE INDEX IF NOT EXISTS idx_promos_end_date ON tbl_promos(end_date);

-- +goose Down
DROP TABLE IF EXISTS tbl_promos;
