
-- +migrate Up
CREATE TABLE tbl_product_category (
	id INTEGER PRIMARY KEY,
	product_id INTEGER NOT NULL,
	category TEXT,
	subcategory TEXT,

	FOREIGN KEY (product_id) REFERENCES tbl_product_category(id)
);

CREATE TABLE tbl_product_specs (
	id INTEGER PRIMARY KEY,
	colours TEXT,
	sizes TEXT,
	segmentation TEXT,
	part_number TEXT,
	power TEXT,
	capacity TEXT,
	scope_of_supply TEXT
);

CREATE TABLE tbl_product (
	id INTEGER PRIMARY KEY,
	serial TEXT NOT NULL,
	name TEXT NOT NULL,
	description TEXT,
	brand TEXT NOT NULL,
	status TEXT NOT NULL,

	product_specs_id INTEGER,

	unit_price_without_vat INTEGER NOT NULL,
	unit_price_with_vat INTEGER NOT NULL,

	unit_price_without_vat_currency TEXT NOT NULL,
	unit_price_with_vat_currency TEXT NOT NULL,

	created_at DATETIME NOT NULL DEFAULT (DATE('1970-01-01 00:00:00')),
	updated_at DATETIME NOT NULL DEFAULT (DATE('1970-01-01 00:00:00')),
	deleted_at DATETIME NOT NULL DEFAULT (DATE('1970-01-01 00:00:00')),

	FOREIGN KEY (product_specs_id) REFERENCES tbl_product_specs(id),
	UNIQUE (serial)
);

CREATE INDEX idx_tbl_product_serial ON tbl_product(serial);
CREATE INDEX idx_tbl_product_brand ON tbl_product(brand);

-- +migrate Down
DROP TABLE tbl_product;
DROP TABLE tbl_product_category;
DROP TABLE tbl_product_specs;
DROP INDEX idx_tbl_product_serial;
DROP INDEX idx_tbl_product_brand;
