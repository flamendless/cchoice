-- +goose Up
-- +goose StatementBegin
CREATE TABLE tbl_checkout_lines (
	id INTEGER PRIMARY KEY,
	checkout_id INTEGER NOT NULL,
	product_id INTEGER NOT NULL,

	name TEXT NOT NULL,
	serial TEXT NOT NULL,
	description TEXT NOT NULL,

	amount INTEGER NOT NULL,
	currency TEXT NOT NULL,
	quantity INTEGER NOT NULL,

	created_at DATETIME NOT NULL DEFAULT (DATE('1970-01-01 00:00:00')),
	updated_at DATETIME NOT NULL DEFAULT (DATE('1970-01-01 00:00:00')),

	FOREIGN KEY (checkout_id) REFERENCES tbl_checkouts(id),
	FOREIGN KEY (product_id) REFERENCES tbl_products(id)
);

CREATE TABLE tbl_checkouts (
	id INTEGER PRIMARY KEY,
	session_id TEXT NOT NULL,

	created_at DATETIME NOT NULL DEFAULT (DATE('1970-01-01 00:00:00')),
	updated_at DATETIME NOT NULL DEFAULT (DATE('1970-01-01 00:00:00'))
);

CREATE TABLE tbl_checkout_payments (
	id TEXT PRIMARY KEY,
	gateway TEXT NOT NULL,
	checkout_id INTEGER NOT NULL,

	status TEXT NOT NULL,
	description TEXT NOT NULL,
	total_amount INTEGER NOT NULL,
	checkout_url TEXT NOT NULL,
	client_key TEXT NOT NULL,
	reference_number TEXT NOT NULL,
	payment_method_type TEXT NOT NULL,
	paid_at DATETIME NOT NULL,
	metadata_remarks TEXT NOT NULL,
	metadata_notes TEXT NOT NULL,
	metadata_customer_number TEXT NOT NULL,

	created_at DATETIME NOT NULL DEFAULT (DATE('1970-01-01 00:00:00')),
	updated_at DATETIME NOT NULL DEFAULT (DATE('1970-01-01 00:00:00')),

	FOREIGN KEY (checkout_id) REFERENCES tbl_checkouts(id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE tbl_checkout_payments;
DROP TABLE tbl_checkouts;
DROP TABLE tbl_checkout_lines;
-- +goose StatementEnd
