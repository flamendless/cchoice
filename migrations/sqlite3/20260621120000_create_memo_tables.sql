-- +goose Up
CREATE TABLE IF NOT EXISTS tbl_memos (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    message TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'DRAFT' CHECK (status IN ('DRAFT', 'PUBLISHED', 'EXPIRED', 'DELETED')),
    start_date TEXT NOT NULL,
    end_date TEXT NOT NULL,
    created_by INTEGER NOT NULL REFERENCES tbl_staffs(id),
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now')),
    deleted_at TEXT NOT NULL DEFAULT '1970-01-01 00:00:00+00:00'
);

CREATE INDEX IF NOT EXISTS idx_memos_status ON tbl_memos(status);
CREATE INDEX IF NOT EXISTS idx_memos_deleted_at ON tbl_memos(deleted_at);
CREATE INDEX IF NOT EXISTS idx_memos_start_date ON tbl_memos(start_date);
CREATE INDEX IF NOT EXISTS idx_memos_end_date ON tbl_memos(end_date);
CREATE INDEX IF NOT EXISTS idx_memos_created_by ON tbl_memos(created_by);

CREATE TABLE IF NOT EXISTS tbl_memo_recipients (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    memo_id INTEGER NOT NULL REFERENCES tbl_memos(id),
    staff_id INTEGER NOT NULL REFERENCES tbl_staffs(id),
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    UNIQUE(memo_id, staff_id)
);

CREATE INDEX IF NOT EXISTS idx_memo_recipients_memo_id ON tbl_memo_recipients(memo_id);
CREATE INDEX IF NOT EXISTS idx_memo_recipients_staff_id ON tbl_memo_recipients(staff_id);

CREATE TABLE IF NOT EXISTS tbl_memo_staff_actions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    memo_id INTEGER NOT NULL REFERENCES tbl_memos(id),
    staff_id INTEGER NOT NULL REFERENCES tbl_staffs(id),
    status TEXT NOT NULL CHECK (status IN ('ACCEPTED', 'REJECTED')),
    reject_reason TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now')),
    accepted_at TEXT,
    rejected_at TEXT,
    UNIQUE(memo_id, staff_id)
);

CREATE INDEX IF NOT EXISTS idx_memo_staff_actions_memo_id ON tbl_memo_staff_actions(memo_id);
CREATE INDEX IF NOT EXISTS idx_memo_staff_actions_staff_id ON tbl_memo_staff_actions(staff_id);

-- +goose Down
DROP TABLE IF EXISTS tbl_memo_staff_actions;
DROP TABLE IF EXISTS tbl_memo_recipients;
DROP TABLE IF EXISTS tbl_memos;
