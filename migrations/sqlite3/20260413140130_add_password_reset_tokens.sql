-- +goose Up
CREATE TABLE IF NOT EXISTS tbl_password_reset_tokens (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    user_type TEXT NOT NULL,
    token_hash TEXT NOT NULL,
    expires_at TEXT NOT NULL,
    used_at TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_user_id_user_type
    ON tbl_password_reset_tokens(user_id, user_type);
CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_token_hash
    ON tbl_password_reset_tokens(token_hash);

-- +goose Down
DROP TABLE IF EXISTS tbl_password_reset_tokens;
