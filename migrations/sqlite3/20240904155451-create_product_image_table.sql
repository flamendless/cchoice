
-- +migrate Up
CREATE TABLE tbl_product_image (
	id INTEGER PRIMARY KEY,
	product_id INTEGER NOT NULL,
	path TEXT NOT NULL,
	thumbnail TEXT,

	created_at DATETIME NOT NULL DEFAULT (DATE('1970-01-01 00:00:00')),
	updated_at DATETIME NOT NULL DEFAULT (DATE('1970-01-01 00:00:00')),
	deleted_at DATETIME NOT NULL DEFAULT (DATE('1970-01-01 00:00:00')),

	FOREIGN KEY (product_id) REFERENCES tbl_product(id)
);

-- +migrate Down
DROP TABLE tbl_product_image;
