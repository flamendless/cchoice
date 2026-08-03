-- +goose Up
ALTER TABLE tbl_invoices ADD COLUMN delivery_date TEXT NOT NULL DEFAULT '';
ALTER TABLE tbl_invoices ADD COLUMN payment_terms_value INTEGER NOT NULL DEFAULT 0;
ALTER TABLE tbl_invoices ADD COLUMN payment_terms_unit TEXT NOT NULL DEFAULT '';

CREATE TABLE IF NOT EXISTS tbl_delivery_receipts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    receipt_number TEXT NOT NULL DEFAULT '',
    invoice_id INTEGER REFERENCES tbl_invoices(id),
    delivered_to TEXT NOT NULL DEFAULT '',
    recipient_email TEXT NOT NULL DEFAULT '',
    tin TEXT NOT NULL DEFAULT '',
    address TEXT NOT NULL DEFAULT '',
    receipt_date TEXT NOT NULL DEFAULT (date('now')),
    terms TEXT NOT NULL DEFAULT '',
    po_number TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'PROCESSING',
    pdf_path TEXT NOT NULL DEFAULT '',
    emailed_at TEXT NOT NULL DEFAULT '',
    created_by INTEGER NOT NULL REFERENCES tbl_staffs(id),
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_delivery_receipts_invoice_id ON tbl_delivery_receipts(invoice_id);
CREATE INDEX IF NOT EXISTS idx_delivery_receipts_status ON tbl_delivery_receipts(status);
CREATE INDEX IF NOT EXISTS idx_delivery_receipts_created_by ON tbl_delivery_receipts(created_by);
CREATE UNIQUE INDEX IF NOT EXISTS idx_delivery_receipts_receipt_number ON tbl_delivery_receipts(receipt_number);

CREATE TABLE IF NOT EXISTS tbl_delivery_receipt_lines (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    delivery_receipt_id INTEGER NOT NULL REFERENCES tbl_delivery_receipts(id),
    quantity TEXT NOT NULL DEFAULT '',
    unit TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_delivery_receipt_lines_receipt_id ON tbl_delivery_receipt_lines(delivery_receipt_id);

CREATE TABLE IF NOT EXISTS tbl_collection_receipts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    receipt_number TEXT NOT NULL DEFAULT '',
    invoice_id INTEGER REFERENCES tbl_invoices(id),
    received_from TEXT NOT NULL DEFAULT '',
    recipient_email TEXT NOT NULL DEFAULT '',
    tin TEXT NOT NULL DEFAULT '',
    address TEXT NOT NULL DEFAULT '',
    receipt_date TEXT NOT NULL DEFAULT (date('now')),
    amount INTEGER NOT NULL DEFAULT 0,
    amount_in_words TEXT NOT NULL DEFAULT '',
    payment_for TEXT NOT NULL DEFAULT '',
    payment_form TEXT NOT NULL DEFAULT 'CASH' CHECK (payment_form IN ('CASH', 'CHECK')),
    sc_citizen_tin TEXT NOT NULL DEFAULT '',
    osca_pwd_id_no TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'PROCESSING',
    pdf_path TEXT NOT NULL DEFAULT '',
    emailed_at TEXT NOT NULL DEFAULT '',
    created_by INTEGER NOT NULL REFERENCES tbl_staffs(id),
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_collection_receipts_invoice_id ON tbl_collection_receipts(invoice_id);
CREATE INDEX IF NOT EXISTS idx_collection_receipts_status ON tbl_collection_receipts(status);
CREATE INDEX IF NOT EXISTS idx_collection_receipts_created_by ON tbl_collection_receipts(created_by);
CREATE UNIQUE INDEX IF NOT EXISTS idx_collection_receipts_receipt_number ON tbl_collection_receipts(receipt_number);

CREATE TABLE IF NOT EXISTS tbl_collection_receipt_settlements (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    collection_receipt_id INTEGER NOT NULL REFERENCES tbl_collection_receipts(id),
    invoice_id INTEGER REFERENCES tbl_invoices(id),
    invoice_number TEXT NOT NULL DEFAULT '',
    amount INTEGER NOT NULL DEFAULT 0,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_collection_receipt_settlements_receipt_id ON tbl_collection_receipt_settlements(collection_receipt_id);

CREATE TABLE IF NOT EXISTS tbl_delivery_receipt_jobs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    queue_id TEXT NOT NULL DEFAULT '',
    delivery_receipt_id INTEGER NOT NULL REFERENCES tbl_delivery_receipts(id),
    job_type TEXT NOT NULL,
    staff_id INTEGER REFERENCES tbl_staffs(id),
    status TEXT NOT NULL DEFAULT 'PENDING',
    error_message TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_delivery_receipt_jobs_receipt_id ON tbl_delivery_receipt_jobs(delivery_receipt_id);
CREATE INDEX IF NOT EXISTS idx_delivery_receipt_jobs_status ON tbl_delivery_receipt_jobs(status);

CREATE TABLE IF NOT EXISTS tbl_collection_receipt_jobs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    queue_id TEXT NOT NULL DEFAULT '',
    collection_receipt_id INTEGER NOT NULL REFERENCES tbl_collection_receipts(id),
    job_type TEXT NOT NULL,
    staff_id INTEGER REFERENCES tbl_staffs(id),
    status TEXT NOT NULL DEFAULT 'PENDING',
    error_message TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_collection_receipt_jobs_receipt_id ON tbl_collection_receipt_jobs(collection_receipt_id);
CREATE INDEX IF NOT EXISTS idx_collection_receipt_jobs_status ON tbl_collection_receipt_jobs(status);

-- +goose Down
DROP TABLE IF EXISTS tbl_collection_receipt_jobs;
DROP TABLE IF EXISTS tbl_delivery_receipt_jobs;
DROP TABLE IF EXISTS tbl_collection_receipt_settlements;
DROP TABLE IF EXISTS tbl_collection_receipts;
DROP TABLE IF EXISTS tbl_delivery_receipt_lines;
DROP TABLE IF EXISTS tbl_delivery_receipts;
