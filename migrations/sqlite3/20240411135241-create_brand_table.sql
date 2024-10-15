
-- +migrate Up
CREATE TABLE tbl_brand (
	id INTEGER PRIMARY KEY,
	name TEXT NOT NULL,

	created_at DATETIME NOT NULL DEFAULT (DATE('1970-01-01 00:00:00')),
	updated_at DATETIME NOT NULL DEFAULT (DATE('1970-01-01 00:00:00')),
	deleted_at DATETIME NOT NULL DEFAULT (DATE('1970-01-01 00:00:00')),

	UNIQUE (name)
);
CREATE INDEX idx_tbl_brand_name ON tbl_brand(name);

-- +migrate Down
DROP INDEX idx_tbl_brand_name;
DROP TABLE tbl_brand;
