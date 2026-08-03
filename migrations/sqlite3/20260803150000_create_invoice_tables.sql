-- +goose Up
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
    proprietor_name TEXT NOT NULL DEFAULT '',
    bir_booklets_info TEXT NOT NULL DEFAULT '',
    bir_authority_to_print TEXT NOT NULL DEFAULT '',
    bir_date_issued TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS tbl_invoice_recipients (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    email TEXT NOT NULL DEFAULT '',
    contact_number TEXT NOT NULL DEFAULT '',
    address TEXT NOT NULL DEFAULT '',
    tin TEXT NOT NULL DEFAULT '',
    registered_name TEXT NOT NULL DEFAULT '',
    notes TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now')),
    deleted_at TEXT NOT NULL DEFAULT '1970-01-01 00:00:00+00:00'
);

CREATE INDEX IF NOT EXISTS idx_invoice_recipients_deleted_at ON tbl_invoice_recipients(deleted_at);
CREATE INDEX IF NOT EXISTS idx_invoice_recipients_name ON tbl_invoice_recipients(name);

CREATE TABLE IF NOT EXISTS tbl_invoices (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    invoice_number TEXT NOT NULL DEFAULT '',
    recipient_id INTEGER REFERENCES tbl_invoice_recipients(id),
    recipient_name TEXT NOT NULL DEFAULT '',
    recipient_email TEXT NOT NULL DEFAULT '',
    recipient_contact_number TEXT NOT NULL DEFAULT '',
    recipient_address TEXT NOT NULL DEFAULT '',
    recipient_tin TEXT NOT NULL DEFAULT '',
    recipient_registered_name TEXT NOT NULL DEFAULT '',
    transaction_type TEXT NOT NULL DEFAULT 'CASH_SALES',
    status TEXT NOT NULL DEFAULT 'PROCESSING',
    issue_date TEXT NOT NULL DEFAULT (date('now')),
    due_date TEXT NOT NULL DEFAULT '',
    notes TEXT NOT NULL DEFAULT '',
    currency TEXT NOT NULL DEFAULT 'PHP',
    subtotal INTEGER NOT NULL DEFAULT 0,
    vat_percentage TEXT NOT NULL DEFAULT '0',
    vat_amount INTEGER NOT NULL DEFAULT 0,
    total INTEGER NOT NULL DEFAULT 0,
    vatable_sales INTEGER NOT NULL DEFAULT 0,
    vat_exempt_sales INTEGER NOT NULL DEFAULT 0,
    zero_rated_sales INTEGER NOT NULL DEFAULT 0,
    total_sales INTEGER NOT NULL DEFAULT 0,
    total_sales_vat_inclusive INTEGER NOT NULL DEFAULT 0,
    less_vat INTEGER NOT NULL DEFAULT 0,
    withholding_tax INTEGER NOT NULL DEFAULT 0,
    amount_net_of_vat INTEGER NOT NULL DEFAULT 0,
    sc_pwd_discount INTEGER NOT NULL DEFAULT 0,
    add_vat INTEGER NOT NULL DEFAULT 0,
    received_amount TEXT NOT NULL DEFAULT '',
    sc_pwd_id_no TEXT NOT NULL DEFAULT '',
    pdf_path TEXT NOT NULL DEFAULT '',
    emailed_at TEXT NOT NULL DEFAULT '',
    created_by INTEGER NOT NULL REFERENCES tbl_staffs(id),
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_invoices_recipient_id ON tbl_invoices(recipient_id);
CREATE INDEX IF NOT EXISTS idx_invoices_status ON tbl_invoices(status);
CREATE INDEX IF NOT EXISTS idx_invoices_created_by ON tbl_invoices(created_by);
CREATE UNIQUE INDEX IF NOT EXISTS idx_invoices_invoice_number ON tbl_invoices(invoice_number);

CREATE TABLE IF NOT EXISTS tbl_invoice_lines (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    invoice_id INTEGER NOT NULL REFERENCES tbl_invoices(id),
    product_id INTEGER REFERENCES tbl_products(id),
    description TEXT NOT NULL DEFAULT '',
    quantity INTEGER NOT NULL DEFAULT 1,
    unit_price INTEGER NOT NULL DEFAULT 0,
    line_total INTEGER NOT NULL DEFAULT 0,
    tax_type TEXT NOT NULL DEFAULT 'VATABLE',
    currency TEXT NOT NULL DEFAULT 'PHP',
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_invoice_lines_invoice_id ON tbl_invoice_lines(invoice_id);
CREATE INDEX IF NOT EXISTS idx_invoice_lines_product_id ON tbl_invoice_lines(product_id);

CREATE TABLE IF NOT EXISTS tbl_invoice_jobs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    queue_id TEXT NOT NULL DEFAULT '',
    invoice_id INTEGER NOT NULL REFERENCES tbl_invoices(id),
    job_type TEXT NOT NULL,
    staff_id INTEGER REFERENCES tbl_staffs(id),
    status TEXT NOT NULL DEFAULT 'PENDING',
    error_message TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_invoice_jobs_invoice_id ON tbl_invoice_jobs(invoice_id);
CREATE INDEX IF NOT EXISTS idx_invoice_jobs_status ON tbl_invoice_jobs(status);

INSERT INTO tbl_invoice_config (id) VALUES (1)
ON CONFLICT(id) DO NOTHING;

-- +goose Down
DROP TABLE IF EXISTS tbl_invoice_jobs;
DROP TABLE IF EXISTS tbl_invoice_lines;
DROP TABLE IF EXISTS tbl_invoices;
DROP TABLE IF EXISTS tbl_invoice_recipients;
DROP TABLE IF EXISTS tbl_invoice_config;
