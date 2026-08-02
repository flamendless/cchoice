-- +goose Up
-- Business/invoice configuration (single row, id = 1)
CREATE TABLE IF NOT EXISTS tbl_invoice_config (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    business_name TEXT NOT NULL DEFAULT '',
    address TEXT NOT NULL DEFAULT '',
    tin TEXT NOT NULL DEFAULT '',
    vat_registration TEXT NOT NULL DEFAULT '',
    email TEXT NOT NULL DEFAULT '',
    contact_number TEXT NOT NULL DEFAULT '',
    website TEXT NOT NULL DEFAULT '',
    footer_notes TEXT NOT NULL DEFAULT '',
    logo_url TEXT NOT NULL DEFAULT '',
    logo_path TEXT NOT NULL DEFAULT '',
    currency TEXT NOT NULL DEFAULT 'PHP',
    vat_percentage TEXT NOT NULL DEFAULT '12',
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

-- Recipients / clients that invoices can be issued to
CREATE TABLE IF NOT EXISTS tbl_invoice_recipients (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    email TEXT NOT NULL DEFAULT '',
    contact_number TEXT NOT NULL DEFAULT '',
    address TEXT NOT NULL DEFAULT '',
    tin TEXT NOT NULL DEFAULT '',
    notes TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now')),
    deleted_at TEXT NOT NULL DEFAULT '1970-01-01 00:00:00+00:00'
);

CREATE INDEX IF NOT EXISTS idx_invoice_recipients_deleted_at ON tbl_invoice_recipients(deleted_at);
CREATE INDEX IF NOT EXISTS idx_invoice_recipients_name ON tbl_invoice_recipients(name);

-- Invoices. Recipient details are snapshotted so historical invoices remain stable.
CREATE TABLE IF NOT EXISTS tbl_invoices (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    invoice_number TEXT NOT NULL DEFAULT '',
    recipient_id INTEGER REFERENCES tbl_invoice_recipients(id),
    recipient_name TEXT NOT NULL DEFAULT '',
    recipient_email TEXT NOT NULL DEFAULT '',
    recipient_contact_number TEXT NOT NULL DEFAULT '',
    recipient_address TEXT NOT NULL DEFAULT '',
    recipient_tin TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'DRAFT',
    issue_date TEXT NOT NULL DEFAULT (date('now')),
    due_date TEXT NOT NULL DEFAULT '',
    notes TEXT NOT NULL DEFAULT '',
    currency TEXT NOT NULL DEFAULT 'PHP',
    subtotal INTEGER NOT NULL DEFAULT 0,
    vat_percentage TEXT NOT NULL DEFAULT '0',
    vat_amount INTEGER NOT NULL DEFAULT 0,
    total INTEGER NOT NULL DEFAULT 0,
    emailed_at TEXT NOT NULL DEFAULT '',
    created_by INTEGER NOT NULL REFERENCES tbl_staffs(id),
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_invoices_recipient_id ON tbl_invoices(recipient_id);
CREATE INDEX IF NOT EXISTS idx_invoices_status ON tbl_invoices(status);
CREATE INDEX IF NOT EXISTS idx_invoices_created_by ON tbl_invoices(created_by);
CREATE UNIQUE INDEX IF NOT EXISTS idx_invoices_invoice_number ON tbl_invoices(invoice_number);

-- Line items belonging to an invoice
CREATE TABLE IF NOT EXISTS tbl_invoice_lines (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    invoice_id INTEGER NOT NULL REFERENCES tbl_invoices(id),
    product_id INTEGER REFERENCES tbl_products(id),
    description TEXT NOT NULL DEFAULT '',
    quantity INTEGER NOT NULL DEFAULT 1,
    unit_price INTEGER NOT NULL DEFAULT 0,
    line_total INTEGER NOT NULL DEFAULT 0,
    currency TEXT NOT NULL DEFAULT 'PHP',
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_invoice_lines_invoice_id ON tbl_invoice_lines(invoice_id);
CREATE INDEX IF NOT EXISTS idx_invoice_lines_product_id ON tbl_invoice_lines(product_id);

-- Seed the singleton config row so updates always target id = 1
INSERT INTO tbl_invoice_config (id) VALUES (1)
ON CONFLICT(id) DO NOTHING;

-- +goose Down
DROP TABLE IF EXISTS tbl_invoice_lines;
DROP TABLE IF EXISTS tbl_invoices;
DROP TABLE IF EXISTS tbl_invoice_recipients;
DROP TABLE IF EXISTS tbl_invoice_config;
