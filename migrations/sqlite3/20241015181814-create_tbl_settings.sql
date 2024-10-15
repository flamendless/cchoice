
-- +migrate Up
CREATE TABLE tbl_settings (
	id INTEGER PRIMARY KEY,
	name TEXT NOT NULL,
	value TEXT NOT NULL,

	UNIQUE (name)
);
CREATE INDEX idx_tbl_settings_name ON tbl_settings(name);

-- +migrate Down
DROP INDEX idx_tbl_settings_name;
DROP TABLE tbl_settings;
