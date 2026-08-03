-- name: CreateCollectionReceipt :one
INSERT INTO tbl_collection_receipts (
    receipt_number,
    invoice_id,
    received_from,
    recipient_email,
    tin,
    address,
    receipt_date,
    amount,
    amount_in_words,
    payment_for,
    payment_form,
    sc_citizen_tin,
    osca_pwd_id_no,
    status,
    pdf_path,
    created_by,
    created_at,
    updated_at
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, '', ?, datetime('now'), datetime('now')
) RETURNING id;

-- name: SetCollectionReceiptNumber :exec
UPDATE tbl_collection_receipts
SET receipt_number = ?, status = 'ISSUED', updated_at = datetime('now')
WHERE id = ?;

-- name: SetCollectionReceiptPDFPath :exec
UPDATE tbl_collection_receipts
SET pdf_path = ?, status = 'ISSUED', updated_at = datetime('now')
WHERE id = ?;

-- name: MarkCollectionReceiptEmailed :exec
UPDATE tbl_collection_receipts
SET emailed_at = datetime('now'), status = 'SENT', updated_at = datetime('now')
WHERE id = ?;

-- name: CreateCollectionReceiptSettlement :exec
INSERT INTO tbl_collection_receipt_settlements (
    collection_receipt_id,
    invoice_id,
    invoice_number,
    amount,
    sort_order,
    created_at,
    updated_at
) VALUES (
    ?, ?, ?, ?, ?, datetime('now'), datetime('now')
);

-- name: GetCollectionReceiptByID :one
SELECT sqlc.embed(tbl_collection_receipts)
FROM tbl_collection_receipts
WHERE id = ?
LIMIT 1;

-- name: GetCollectionReceiptSettlementsByReceiptID :many
SELECT sqlc.embed(tbl_collection_receipt_settlements)
FROM tbl_collection_receipt_settlements
WHERE collection_receipt_id = ?
ORDER BY sort_order ASC, id ASC;

-- name: CountCollectionReceipts :one
SELECT COUNT(*) AS count FROM tbl_collection_receipts;

-- name: ListCollectionReceiptsPaginated :many
SELECT
    tbl_collection_receipts.id,
    tbl_collection_receipts.receipt_number,
    tbl_collection_receipts.received_from,
    tbl_collection_receipts.recipient_email,
    tbl_collection_receipts.status,
    tbl_collection_receipts.receipt_date,
    tbl_collection_receipts.amount,
    tbl_collection_receipts.pdf_path,
    tbl_collection_receipts.emailed_at,
    tbl_collection_receipts.created_at
FROM tbl_collection_receipts
ORDER BY tbl_collection_receipts.id DESC
LIMIT @limit OFFSET @offset;
