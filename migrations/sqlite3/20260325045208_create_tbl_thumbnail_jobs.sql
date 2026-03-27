
-- +goose Up
CREATE TABLE tbl_thumbnail_jobs (
	id INTEGER PRIMARY KEY,
	queue_id TEXT NOT NULL,
	product_id INTEGER NOT NULL,
	brand TEXT NOT NULL,
	source_path TEXT NOT NULL,
	status TEXT NOT NULL DEFAULT 'pending',
	error_message TEXT,

	created_at DATETIME NOT NULL DEFAULT (DATE('1970-01-01 00:00:00')),
	updated_at DATETIME NOT NULL DEFAULT (DATE('1970-01-01 00:00:00')),

	FOREIGN KEY (product_id) REFERENCES tbl_products(id)
);

CREATE INDEX idx_tbl_thumbnail_jobs_product_id ON tbl_thumbnail_jobs(product_id);
CREATE INDEX idx_tbl_thumbnail_jobs_status ON tbl_thumbnail_jobs(status);

-- +goose Down
DROP INDEX idx_tbl_thumbnail_jobs_product_id;
DROP INDEX idx_tbl_thumbnail_jobs_status;
DROP TABLE tbl_thumbnail_jobs;
