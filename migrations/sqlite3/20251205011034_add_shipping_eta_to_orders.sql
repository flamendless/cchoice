-- +goose Up
-- +goose StatementBegin
ALTER TABLE tbl_orders ADD COLUMN shipping_eta TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE tbl_orders DROP COLUMN shipping_eta;
-- +goose StatementEnd

