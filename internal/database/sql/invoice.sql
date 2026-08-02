-- name: GetInvoiceConfig :one
SELECT sqlc.embed(tbl_invoice_config)
FROM tbl_invoice_config
WHERE id = 1
LIMIT 1;

-- name: UpsertInvoiceConfig :exec
INSERT INTO tbl_invoice_config (
    id,
    business_name,
    address,
    tin,
    vat_registration,
    email,
    contact_number,
    website,
    footer_notes,
    logo_url,
    logo_path,
    currency,
    vat_percentage,
    created_at,
    updated_at
) VALUES (
    1, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now'), datetime('now')
)
ON CONFLICT(id) DO UPDATE SET
    business_name = excluded.business_name,
    address = excluded.address,
    tin = excluded.tin,
    vat_registration = excluded.vat_registration,
    email = excluded.email,
    contact_number = excluded.contact_number,
    website = excluded.website,
    footer_notes = excluded.footer_notes,
    logo_url = excluded.logo_url,
    logo_path = excluded.logo_path,
    currency = excluded.currency,
    vat_percentage = excluded.vat_percentage,
    updated_at = datetime('now');

-- name: GetAllInvoiceRecipients :many
SELECT sqlc.embed(tbl_invoice_recipients)
FROM tbl_invoice_recipients
WHERE deleted_at = '1970-01-01 00:00:00+00:00'
    AND (@search IS NULL OR @search = ''
        OR LOWER(name) LIKE '%' || LOWER(@search) || '%'
        OR LOWER(email) LIKE '%' || LOWER(@search) || '%')
ORDER BY name ASC;

-- name: GetInvoiceRecipientByID :one
SELECT sqlc.embed(tbl_invoice_recipients)
FROM tbl_invoice_recipients
WHERE id = ?
    AND deleted_at = '1970-01-01 00:00:00+00:00'
LIMIT 1;

-- name: CreateInvoiceRecipient :one
INSERT INTO tbl_invoice_recipients (
    name,
    email,
    contact_number,
    address,
    tin,
    notes,
    created_at,
    updated_at
) VALUES (
    ?, ?, ?, ?, ?, ?, datetime('now'), datetime('now')
) RETURNING id;

-- name: UpdateInvoiceRecipient :exec
UPDATE tbl_invoice_recipients
SET
    name = ?,
    email = ?,
    contact_number = ?,
    address = ?,
    tin = ?,
    notes = ?,
    updated_at = datetime('now')
WHERE id = ?
    AND deleted_at = '1970-01-01 00:00:00+00:00';

-- name: SoftDeleteInvoiceRecipient :exec
UPDATE tbl_invoice_recipients
SET deleted_at = datetime('now')
WHERE id = ?;

-- name: CreateInvoice :one
INSERT INTO tbl_invoices (
    invoice_number,
    recipient_id,
    recipient_name,
    recipient_email,
    recipient_contact_number,
    recipient_address,
    recipient_tin,
    status,
    issue_date,
    due_date,
    notes,
    currency,
    subtotal,
    vat_percentage,
    vat_amount,
    total,
    created_by,
    created_at,
    updated_at
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now'), datetime('now')
) RETURNING id;

-- name: SetInvoiceNumber :exec
UPDATE tbl_invoices
SET invoice_number = ?, updated_at = datetime('now')
WHERE id = ?;

-- name: CreateInvoiceLine :exec
INSERT INTO tbl_invoice_lines (
    invoice_id,
    product_id,
    description,
    quantity,
    unit_price,
    line_total,
    currency,
    created_at,
    updated_at
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, datetime('now'), datetime('now')
);

-- name: GetInvoiceByID :one
SELECT sqlc.embed(tbl_invoices)
FROM tbl_invoices
WHERE id = ?
LIMIT 1;

-- name: GetInvoiceLinesByInvoiceID :many
SELECT sqlc.embed(tbl_invoice_lines)
FROM tbl_invoice_lines
WHERE invoice_id = ?
ORDER BY id ASC;

-- name: GetAllInvoices :many
SELECT
    tbl_invoices.id,
    tbl_invoices.invoice_number,
    tbl_invoices.recipient_name,
    tbl_invoices.recipient_email,
    tbl_invoices.status,
    tbl_invoices.issue_date,
    tbl_invoices.currency,
    tbl_invoices.total,
    tbl_invoices.emailed_at,
    tbl_invoices.created_at
FROM tbl_invoices
ORDER BY tbl_invoices.id DESC
LIMIT 500;

-- name: UpdateInvoiceStatus :exec
UPDATE tbl_invoices
SET status = ?, updated_at = datetime('now')
WHERE id = ?;

-- name: MarkInvoiceEmailed :exec
UPDATE tbl_invoices
SET emailed_at = datetime('now'), status = 'SENT', updated_at = datetime('now')
WHERE id = ?;
