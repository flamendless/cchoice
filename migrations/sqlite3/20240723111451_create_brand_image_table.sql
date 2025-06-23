
-- +goose Up
CREATE TABLE tbl_brand_images (
	id INTEGER PRIMARY KEY,
	brand_id INTEGER NOT NULL,
	path TEXT NOT NULL,
	is_main boolean NOT NULL DEFAULT false,

	created_at DATETIME NOT NULL DEFAULT (DATE('1970-01-01 00:00:00')),
	updated_at DATETIME NOT NULL DEFAULT (DATE('1970-01-01 00:00:00')),
	deleted_at DATETIME NOT NULL DEFAULT (DATE('1970-01-01 00:00:00')),

	FOREIGN KEY (brand_id) REFERENCES tbl_brands(id)
);

-- +goose Down
DROP TABLE tbl_brand_images;
