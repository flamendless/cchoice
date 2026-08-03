-- sqlc schema stub: goose creates this table at runtime; not applied via project migrations.
CREATE TABLE goose_db_version (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    version_id INTEGER NOT NULL,
    is_applied INTEGER NOT NULL,
    tstamp TIMESTAMP DEFAULT (datetime('now'))
);
