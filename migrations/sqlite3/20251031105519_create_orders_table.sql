-- +goose Up
-- +goose StatementBegin
CREATE TABLE tbl_orders (
	id INTEGER PRIMARY KEY,
	checkout_id INTEGER NOT NULL,
	checkout_payment_id TEXT NOT NULL,

	order_number TEXT NOT NULL UNIQUE,
	status TEXT NOT NULL CHECK (status IN ('PENDING', 'CONFIRMED', 'PROCESSING', 'SHIPPED', 'DELIVERED', 'CANCELLED', 'REFUNDED')),

	customer_name TEXT NOT NULL,
	customer_email TEXT NOT NULL,
	customer_phone TEXT NOT NULL,

	billing_address_line1 TEXT NOT NULL,
	billing_address_line2 TEXT NOT NULL,
	billing_city TEXT NOT NULL,
	billing_state TEXT NOT NULL,
	billing_postal_code TEXT NOT NULL,
	billing_country TEXT NOT NULL DEFAULT 'PH',
	billing_latitude TEXT,
	billing_longitude TEXT,
	billing_formatted_address TEXT,
	billing_place_id TEXT,

	shipping_address_line1 TEXT NOT NULL,
	shipping_address_line2 TEXT NOT NULL,
	shipping_city TEXT NOT NULL,
	shipping_state TEXT NOT NULL,
	shipping_postal_code TEXT NOT NULL,
	shipping_country TEXT NOT NULL DEFAULT 'PH',
	shipping_latitude TEXT,
	shipping_longitude TEXT,
	shipping_formatted_address TEXT,
	shipping_place_id TEXT,

	subtotal_amount INTEGER NOT NULL,
	shipping_amount INTEGER NOT NULL,
	discount_amount INTEGER NOT NULL DEFAULT 0,
	total_amount INTEGER NOT NULL,
	currency TEXT NOT NULL DEFAULT 'PHP',

	shipping_service TEXT,
	shipping_order_id TEXT,
	shipping_tracking_number TEXT,

	notes TEXT,
	remarks TEXT,

	created_at DATETIME NOT NULL DEFAULT (DATE('1970-01-01 00:00:00')),
	updated_at DATETIME NOT NULL DEFAULT (DATE('1970-01-01 00:00:00')),

	FOREIGN KEY (checkout_id) REFERENCES tbl_checkouts(id),
	FOREIGN KEY (checkout_payment_id) REFERENCES tbl_checkout_payments(id)
);

CREATE TABLE tbl_order_lines (
	id INTEGER PRIMARY KEY,
	order_id INTEGER NOT NULL,
	checkout_line_id INTEGER NOT NULL,
	product_id INTEGER NOT NULL,

	name TEXT NOT NULL,
	serial TEXT NOT NULL,
	description TEXT NOT NULL,

	unit_price INTEGER NOT NULL,
	quantity INTEGER NOT NULL,
	total_price INTEGER NOT NULL,
	currency TEXT NOT NULL,

	created_at DATETIME NOT NULL DEFAULT (DATE('1970-01-01 00:00:00')),
	updated_at DATETIME NOT NULL DEFAULT (DATE('1970-01-01 00:00:00')),

	FOREIGN KEY (order_id) REFERENCES tbl_orders(id) ON DELETE CASCADE,
	FOREIGN KEY (checkout_line_id) REFERENCES tbl_checkout_lines(id),
	FOREIGN KEY (product_id) REFERENCES tbl_products(id)
);

CREATE INDEX idx_orders_checkout_id ON tbl_orders(checkout_id);
CREATE INDEX idx_orders_checkout_payment_id ON tbl_orders(checkout_payment_id);
CREATE INDEX idx_orders_order_number ON tbl_orders(order_number);
CREATE INDEX idx_orders_status ON tbl_orders(status);
CREATE INDEX idx_orders_customer_email ON tbl_orders(customer_email);
CREATE INDEX idx_orders_created_at ON tbl_orders(created_at);

CREATE INDEX idx_order_lines_order_id ON tbl_order_lines(order_id);
CREATE INDEX idx_order_lines_product_id ON tbl_order_lines(product_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX idx_order_lines_product_id;
DROP INDEX idx_order_lines_order_id;
DROP INDEX idx_orders_created_at;
DROP INDEX idx_orders_customer_email;
DROP INDEX idx_orders_status;
DROP INDEX idx_orders_order_number;
DROP INDEX idx_orders_checkout_payment_id;
DROP INDEX idx_orders_checkout_id;

DROP TABLE tbl_order_lines;
DROP TABLE tbl_orders;
-- +goose StatementEnd