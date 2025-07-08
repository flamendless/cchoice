-- +goose Up
-- +goose StatementBegin
CREATE TABLE tbl_checkout_line_items (
	id INTEGER PRIMARY KEY,
	checkout_id TEXT NOT NULL,

	amount INTEGER NOT NULL,
	currency TEXT NOT NULL,
	description TEXT NOT NULL,
	name TEXT NOT NULL,
	quantity INTEGER NOT NULL,

	created_at DATETIME NOT NULL DEFAULT (DATE('1970-01-01 00:00:00')),
	updated_at DATETIME NOT NULL DEFAULT (DATE('1970-01-01 00:00:00')),

	FOREIGN KEY (checkout_id) REFERENCES tbl_checkouts(id)
);

CREATE TABLE tbl_checkouts (
	id TEXT PRIMARY KEY,
	gateway TEXT NOT NULL,

	status TEXT NOT NULL,
	description TEXT NOT NULL,
	total_amount INTEGER NOT NULL,
	checkout_url TEXT NOT NULL,
	client_key TEXT NOT NULL,
	reference_number TEXT NOT NULL,
	payment_status TEXT NOT NULL,
	payment_method_type TEXT NOT NULL,
	paid_at DATETIME NOT NULL,
	metadata_remarks TEXT NOT NULL,
	metadata_notes TEXT NOT NULL,
	metadata_customer_number TEXT NOT NULL,

	created_at DATETIME NOT NULL DEFAULT (DATE('1970-01-01 00:00:00')),
	updated_at DATETIME NOT NULL DEFAULT (DATE('1970-01-01 00:00:00'))
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE tbl_checkouts;
DROP TABLE tbl_checkout_line_items;
-- +goose StatementEnd
