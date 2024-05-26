CREATE TABLE tbl_product_category (
	id INTEGER PRIMARY KEY,
	category TEXT,
	subcategory TEXT
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

	product_category_id INTEGER,
	product_sepcs_id INTEGER,

	unit_price_without_vat INTEGER NOT NULL,
	unit_price_with_vat INTEGER NOT NULL,

	unit_price_without_vat_currency TEXT NOT NULL,
	unit_price_with_vat_currency TEXT NOT NULL,

	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL,
	deleted_at TEXT NOT NULL,

	FOREIGN KEY (product_category_id) REFERENCES tbl_product_category(id),
	FOREIGN KEY (product_specs_id) REFERENCES tbl_product_specs(id),
	UNIQUE (serial)
);
