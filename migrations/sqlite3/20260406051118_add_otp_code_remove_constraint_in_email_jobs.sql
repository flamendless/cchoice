-- +goose Up
-- Remove CHECK constraint by recreating the table

ALTER TABLE tbl_email_jobs RENAME TO tbl_email_jobs_old;

CREATE TABLE tbl_email_jobs (
	id INTEGER PRIMARY KEY,
	queue_id TEXT NOT NULL,
	recipient TEXT NOT NULL,
	cc TEXT,
	subject TEXT NOT NULL,
	template_name TEXT NOT NULL,
	order_id INTEGER,
	otp_code TEXT,
	checkout_payment_id TEXT,
	created_at DATETIME NOT NULL DEFAULT (DATETIME('now')),
	updated_at DATETIME NOT NULL DEFAULT (DATETIME('now')),

	FOREIGN KEY (order_id) REFERENCES tbl_orders(id),
	FOREIGN KEY (checkout_payment_id) REFERENCES tbl_checkout_payments(id)
);

INSERT INTO tbl_email_jobs (
	id,
	queue_id,
	recipient,
	cc,
	subject,
	template_name,
	order_id,
	checkout_payment_id,
	created_at,
	updated_at
)
SELECT
	id,
	queue_id,
	recipient,
	cc,
	subject,
	template_name,
	order_id,
	checkout_payment_id,
	created_at,
	updated_at
FROM tbl_email_jobs_old;

DROP TABLE tbl_email_jobs_old;


-- +goose Down
-- Re-add CHECK constraint

ALTER TABLE tbl_email_jobs RENAME TO tbl_email_jobs_old;

CREATE TABLE tbl_email_jobs (
	id INTEGER PRIMARY KEY,
	queue_id TEXT NOT NULL,
	recipient TEXT NOT NULL,
	cc TEXT,
	subject TEXT NOT NULL,
	template_name TEXT NOT NULL CHECK (
		template_name IN ('order_confirmation', 'payment_confirmation')
	),
	order_id INTEGER,
	checkout_payment_id TEXT,
	created_at DATETIME NOT NULL DEFAULT (DATETIME('now')),
	updated_at DATETIME NOT NULL DEFAULT (DATETIME('now')),

	FOREIGN KEY (order_id) REFERENCES tbl_orders(id),
	FOREIGN KEY (checkout_payment_id) REFERENCES tbl_checkout_payments(id)
);

INSERT INTO tbl_email_jobs (
	id,
	queue_id,
	recipient,
	cc,
	subject,
	template_name,
	order_id,
	checkout_payment_id,
	created_at,
	updated_at
)
SELECT
	id,
	queue_id,
	recipient,
	cc,
	subject,
	template_name,
	order_id,
	checkout_payment_id,
	created_at,
	updated_at
FROM tbl_email_jobs_old;

DROP TABLE tbl_email_jobs_old;
