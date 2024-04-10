CREATE TABLE product_category (
	id INTEGER PRIMARY KEY,
	category TEXT,
	subcategory TEXT
);

CREATE TABLE product_type (
	id INTEGER PRIMARY KEY,
	colours TEXT,
	sizes TEXT,
	segmentation TEXT
);

CREATE TABLE product (
	id INTEGER PRIMARY KEY,
	name TEXT NOT NULL,
	description TEXT,

	product_category_id INTEGER,
	product_type_id INTEGER,

	unit_price_without_vat INTEGER NOT NULL,
	unit_price_with_vat INTEGER NOT NULL,

	unit_price_without_vat_currency TEXT NOT NULL,
	unit_price_with_vat_currency TEXT NOT NULL,

	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL,
	deleted_at TEXT NOT NULL,

	FOREIGN KEY (product_category_id) REFERENCES product_category(id),
	FOREIGN KEY (product_type_id) REFERENCES product_type(id)
);
