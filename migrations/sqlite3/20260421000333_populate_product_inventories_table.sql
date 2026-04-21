-- +goose Up
-- +goose StatementBegin
INSERT OR IGNORE INTO tbl_product_inventories(
    product_id,
    stocks,
    stocks_in
)
SELECT id, 0, 'SUPPLIER'
FROM tbl_products;
-- +goose StatementEnd

-- +goose Down
