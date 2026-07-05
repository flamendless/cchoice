-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS tbl_order_status_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    order_id INTEGER NOT NULL REFERENCES tbl_orders(id),
    staff_id INTEGER REFERENCES tbl_staffs(id),
    from_status TEXT CHECK (from_status IS NULL OR from_status IN ('PENDING', 'CONFIRMED', 'PROCESSING', 'SHIPPED', 'DELIVERED', 'CANCELLED', 'REFUNDED')),
    to_status TEXT NOT NULL CHECK (to_status IN ('PENDING', 'CONFIRMED', 'PROCESSING', 'SHIPPED', 'DELIVERED', 'CANCELLED', 'REFUNDED')),
    notes TEXT,
    created_at DATETIME NOT NULL DEFAULT (datetime('now')),
    updated_at DATETIME NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_order_status_history_order_id ON tbl_order_status_history(order_id);
CREATE INDEX IF NOT EXISTS idx_order_status_history_created_at ON tbl_order_status_history(created_at);

INSERT INTO tbl_order_status_history (order_id, staff_id, from_status, to_status, notes, created_at, updated_at)
SELECT
    id,
    NULL,
    NULL,
    status,
    NULL,
    created_at,
    created_at
FROM tbl_orders;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_order_status_history_created_at;
DROP INDEX IF EXISTS idx_order_status_history_order_id;
DROP TABLE IF EXISTS tbl_order_status_history;
-- +goose StatementEnd
