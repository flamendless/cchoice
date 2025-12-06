-- +goose Up
-- +goose StatementBegin
CREATE TABLE tbl_email_jobs (
	id INTEGER PRIMARY KEY,
	queue_id TEXT NOT NULL,
	recipient TEXT NOT NULL,
	cc TEXT,
	subject TEXT NOT NULL,
	template_name TEXT NOT NULL CHECK (template_name IN ('order_confirmation', 'payment_confirmation')),
	order_id INTEGER,
	checkout_payment_id TEXT,
	created_at DATETIME NOT NULL DEFAULT (DATETIME('now')),
	updated_at DATETIME NOT NULL DEFAULT (DATETIME('now')),

	FOREIGN KEY (order_id) REFERENCES tbl_orders(id),
	FOREIGN KEY (checkout_payment_id) REFERENCES tbl_checkout_payments(id)
);

CREATE INDEX idx_email_jobs_queue_id ON tbl_email_jobs(queue_id);
CREATE INDEX idx_email_jobs_template_name ON tbl_email_jobs(template_name);
CREATE INDEX idx_email_jobs_order_id ON tbl_email_jobs(order_id);
CREATE INDEX idx_email_jobs_checkout_payment_id ON tbl_email_jobs(checkout_payment_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX idx_email_jobs_checkout_payment_id;
DROP INDEX idx_email_jobs_order_id;
DROP INDEX idx_email_jobs_template_name;
DROP INDEX idx_email_jobs_queue_id;
DROP TABLE tbl_email_jobs;
-- +goose StatementEnd
