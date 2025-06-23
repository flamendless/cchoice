
-- +goose Up
CREATE TABLE tbl_brands (
	id INTEGER PRIMARY KEY,
	name TEXT NOT NULL,

	created_at DATETIME NOT NULL DEFAULT (DATE('1970-01-01 00:00:00')),
	updated_at DATETIME NOT NULL DEFAULT (DATE('1970-01-01 00:00:00')),
	deleted_at DATETIME NOT NULL DEFAULT (DATE('1970-01-01 00:00:00')),

	UNIQUE (name)
);
CREATE INDEX idx_tbl_brands_name ON tbl_brands(name);

-- +goose Down
DROP INDEX idx_tbl_brands_name;
DROP TABLE tbl_brands;
