-- +goose Up
-- +goose StatementBegin
CREATE TABLE tbl_tracked_links (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    slug TEXT UNIQUE NOT NULL,
    destination_url TEXT NOT NULL,

    source TEXT,
    medium TEXT,
    campaign TEXT,

    status TEXT NOT NULL DEFAULT 'DRAFT',
    staff_id TEXT,

    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE tbl_link_clicks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    link_id TEXT NOT NULL,
    clicked_at TEXT NOT NULL DEFAULT (datetime('now')),

    referrer TEXT,
    user_agent TEXT,
    ip_hash TEXT,
    device TEXT,

    utm_source TEXT,
    utm_medium TEXT,
    utm_campaign TEXT,

    FOREIGN KEY (link_id) REFERENCES tbl_tracked_links(id)
);

CREATE INDEX idx_link_clicks_link_id ON tbl_link_clicks(link_id);
CREATE INDEX idx_link_clicks_clicked_at ON tbl_link_clicks(clicked_at);
-- +goose StatementEnd

-- +goose Down
DROP INDEX IF EXISTS idx_link_clicks_link_id;
DROP INDEX IF EXISTS idx_link_clicks_clicked_at;
DROP TABLE IF EXISTS tbl_link_clicks;
DROP TABLE IF EXISTS tbl_tracked_links;
