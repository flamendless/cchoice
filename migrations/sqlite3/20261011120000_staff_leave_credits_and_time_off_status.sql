-- +goose Up
-- +goose StatementBegin
ALTER TABLE tbl_staff_time_offs ADD COLUMN hours REAL NOT NULL DEFAULT 0 CHECK (hours >= 0);
ALTER TABLE tbl_staff_time_offs ADD COLUMN status TEXT NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'APPROVED', 'REJECTED'));
ALTER TABLE tbl_staff_time_offs ADD COLUMN reject_reason TEXT;
ALTER TABLE tbl_staff_time_offs ADD COLUMN rejected_by INTEGER REFERENCES tbl_staffs(id);
ALTER TABLE tbl_staff_time_offs ADD COLUMN rejected_at DATETIME;

UPDATE tbl_staff_time_offs SET status = 'APPROVED' WHERE approved = 1;
UPDATE tbl_staff_time_offs SET status = 'PENDING' WHERE approved IS NULL OR approved = 0;

CREATE TABLE IF NOT EXISTS tbl_staff_leave_credit_grants (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    staff_id INTEGER NOT NULL REFERENCES tbl_staffs(id),
    hours REAL NOT NULL CHECK (hours > 0),
    note TEXT,
    granted_by INTEGER NOT NULL REFERENCES tbl_staffs(id),
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_staff_leave_credit_grants_staff_id ON tbl_staff_leave_credit_grants(staff_id);
CREATE INDEX IF NOT EXISTS idx_staff_time_offs_status ON tbl_staff_time_offs(status);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_staff_time_offs_status;
DROP INDEX IF EXISTS idx_staff_leave_credit_grants_staff_id;
DROP TABLE IF EXISTS tbl_staff_leave_credit_grants;
-- SQLite down migration cannot easily drop added columns on tbl_staff_time_offs without table rebuild; leave as forward-only data shape.
-- +goose StatementEnd
