
-- +goose Up
CREATE TABLE tbl_product_categories (
	id INTEGER PRIMARY KEY,
	category TEXT,
	subcategory TEXT,
	promoted_at_homepage bool DEFAULT false
);
CREATE INDEX idx_tbl_product_categories_category ON tbl_product_categories(category);

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

CREATE TABLE tbl_products (
	id INTEGER PRIMARY KEY,
	serial TEXT NOT NULL,
	name TEXT NOT NULL,
	description TEXT,
	brand_id INTEGER NOT NULL,
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
	FOREIGN KEY (brand_id) REFERENCES tbl_brand(id),
	UNIQUE (serial)
);

CREATE INDEX idx_tbl_products_serial ON tbl_products(serial);

CREATE TABLE tbl_products_categories (
	id INTEGER PRIMARY KEY,
	category_id INTEGER NOT NULL,
	product_id INTEGER NOT NULL,
	FOREIGN KEY (category_id) REFERENCES tbl_product_categories(id),
	FOREIGN KEY (product_id) REFERENCES tbl_products(id),
	UNIQUE(product_id, category_id)
);

-- +goose Down
DROP INDEX idx_tbl_product_categories_category;
DROP INDEX idx_tbl_products_serial;
DROP TABLE tbl_products_categories;
DROP TABLE tbl_product_specs;
DROP TABLE tbl_product_categories;
DROP TABLE tbl_products;
