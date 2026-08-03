-- name: CreateDeliveryReceipt :one
INSERT INTO tbl_delivery_receipts (
    receipt_number,
    invoice_id,
    delivered_to,
    recipient_email,
    tin,
    address,
    receipt_date,
    terms,
    po_number,
    status,
    pdf_path,
    created_by,
    created_at,
    updated_at
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, '', ?, datetime('now'), datetime('now')
) RETURNING id;

-- name: SetDeliveryReceiptNumber :exec
UPDATE tbl_delivery_receipts
SET receipt_number = ?, status = 'ISSUED', updated_at = datetime('now')
WHERE id = ?;

-- name: SetDeliveryReceiptPDFPath :exec
UPDATE tbl_delivery_receipts
SET pdf_path = ?, status = 'ISSUED', updated_at = datetime('now')
WHERE id = ?;

-- name: MarkDeliveryReceiptEmailed :exec
UPDATE tbl_delivery_receipts
SET emailed_at = datetime('now'), status = 'SENT', updated_at = datetime('now')
WHERE id = ?;

-- name: CreateDeliveryReceiptLine :exec
INSERT INTO tbl_delivery_receipt_lines (
    delivery_receipt_id,
    quantity,
    unit,
    description,
    sort_order,
    created_at,
    updated_at
) VALUES (
    ?, ?, ?, ?, ?, datetime('now'), datetime('now')
);

-- name: GetDeliveryReceiptByID :one
SELECT sqlc.embed(tbl_delivery_receipts)
FROM tbl_delivery_receipts
WHERE id = ?
LIMIT 1;

-- name: GetDeliveryReceiptLinesByReceiptID :many
SELECT sqlc.embed(tbl_delivery_receipt_lines)
FROM tbl_delivery_receipt_lines
WHERE delivery_receipt_id = ?
ORDER BY sort_order ASC, id ASC;

-- name: CountDeliveryReceipts :one
SELECT COUNT(*) AS count FROM tbl_delivery_receipts;

-- name: ListDeliveryReceiptsPaginated :many
SELECT
    tbl_delivery_receipts.id,
    tbl_delivery_receipts.receipt_number,
    tbl_delivery_receipts.delivered_to,
    tbl_delivery_receipts.recipient_email,
    tbl_delivery_receipts.status,
    tbl_delivery_receipts.receipt_date,
    tbl_delivery_receipts.pdf_path,
    tbl_delivery_receipts.emailed_at,
    tbl_delivery_receipts.created_at
FROM tbl_delivery_receipts
ORDER BY tbl_delivery_receipts.id DESC
LIMIT @limit OFFSET @offset;
