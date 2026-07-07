-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS tbl_quotation_status_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    quotation_id INTEGER NOT NULL REFERENCES tbl_quotations(id),
    staff_id INTEGER REFERENCES tbl_staffs(id),
    from_status TEXT,
    to_status TEXT NOT NULL,
    notes TEXT,
    created_at DATETIME NOT NULL DEFAULT (datetime('now')),
    updated_at DATETIME NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_quotation_status_history_quotation_id ON tbl_quotation_status_history(quotation_id);
CREATE INDEX IF NOT EXISTS idx_quotation_status_history_created_at ON tbl_quotation_status_history(created_at);

INSERT INTO tbl_quotation_status_history (quotation_id, staff_id, from_status, to_status, notes, created_at, updated_at)
SELECT
    id,
    NULL,
    NULL,
    status,
    NULL,
    created_at,
    created_at
FROM tbl_quotations
WHERE status != 'DRAFT';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_quotation_status_history_created_at;
DROP INDEX IF EXISTS idx_quotation_status_history_quotation_id;
DROP TABLE IF EXISTS tbl_quotation_status_history;
-- +goose StatementEnd
