-- +goose Up
CREATE TABLE IF NOT EXISTS tbl_useragents (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_agent TEXT NOT NULL UNIQUE,
    browser TEXT NOT NULL DEFAULT '',
    browser_version TEXT NOT NULL DEFAULT '',
    os TEXT NOT NULL DEFAULT '',
    device TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

ALTER TABLE tbl_staff_accesses DROP COLUMN user_agent;
ALTER TABLE tbl_staff_attendances ADD COLUMN useragent_id INTEGER REFERENCES tbl_useragents(id);
ALTER TABLE tbl_staff_accesses ADD COLUMN useragent_id INTEGER REFERENCES tbl_useragents(id);

-- +goose Down
ALTER TABLE tbl_staff_accesses DROP COLUMN useragent_id;
ALTER TABLE tbl_staff_attendances DROP COLUMN useragent_id;
DROP TABLE IF EXISTS tbl_useragents;
