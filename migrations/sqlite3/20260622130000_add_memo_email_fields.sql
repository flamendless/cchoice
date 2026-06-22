-- +goose Up
ALTER TABLE tbl_memos ADD COLUMN emails_sent_at TEXT NOT NULL DEFAULT '1970-01-01 00:00:00+00:00';

ALTER TABLE tbl_email_jobs ADD COLUMN memo_id INTEGER REFERENCES tbl_memos(id);

-- +goose Down
