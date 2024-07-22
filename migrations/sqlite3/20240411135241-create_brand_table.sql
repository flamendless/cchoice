
-- +migrate Up
CREATE TABLE tbl_brand (
	id INTEGER PRIMARY KEY,
	name TEXT NOT NULL,

	UNIQUE (name)
);
CREATE INDEX idx_tbl_brand_name ON tbl_brand(name);

-- +migrate Down
DROP TABLE tbl_brand;
DROP INDEX idx_tbl_brand_name;
