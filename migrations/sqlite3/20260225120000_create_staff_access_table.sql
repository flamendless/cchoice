-- +goose Up
CREATE TABLE IF NOT EXISTS tbl_staff_accesses (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    staff_id INTEGER NOT NULL,
    login_at TEXT NOT NULL,
    logout_at TEXT,
    user_agent TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (staff_id) REFERENCES tbl_staffs(id)
);

CREATE INDEX IF NOT EXISTS idx_staff_accesses_staff_id ON tbl_staff_accesses(staff_id);
CREATE INDEX IF NOT EXISTS idx_staff_accesses_login_at ON tbl_staff_accesses(login_at);

-- +goose Down
DROP INDEX IF EXISTS idx_staff_accesses_login_at;
DROP INDEX IF EXISTS idx_staff_accesses_staff_id;
DROP TABLE IF EXISTS tbl_staff_accesses;
