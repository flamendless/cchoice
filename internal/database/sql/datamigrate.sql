-- name: ListDatamigrateAppliedNames :many
SELECT name FROM tbl_datamigrate_applied ORDER BY name;

-- name: UpsertDatamigrateApplied :exec
INSERT INTO tbl_datamigrate_applied (name, applied_at)
VALUES (?, datetime('now'))
ON CONFLICT(name) DO NOTHING;

-- name: GetGooseMaxAppliedVersion :one
SELECT CAST(COALESCE(MAX(version_id), 0) AS INTEGER) AS max_version FROM goose_db_version WHERE is_applied = 1;

-- name: ListGooseAppliedVersionIDs :many
SELECT version_id FROM goose_db_version WHERE is_applied = 1 ORDER BY version_id;
