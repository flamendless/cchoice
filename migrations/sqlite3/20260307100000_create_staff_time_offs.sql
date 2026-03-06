-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS tbl_staff_time_offs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    type TEXT NOT NULL CHECK (type IN ('VL', 'SL', 'ABSENT')),
    start_date DATETIME NOT NULL,
    end_date DATETIME NOT NULL,
    staff_id INTEGER NOT NULL REFERENCES tbl_staffs(id),
    useragent_id INTEGER REFERENCES tbl_useragents(id),
    description TEXT NOT NULL,
    approved BOOLEAN DEFAULT FALSE,
    approved_by INTEGER REFERENCES tbl_staffs(id),
    approved_at DATETIME,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_staff_time_offs_type ON tbl_staff_time_offs(type);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_staff_time_offs_type;
DROP TABLE IF EXISTS tbl_staff_time_offs;
-- +goose StatementEnd
