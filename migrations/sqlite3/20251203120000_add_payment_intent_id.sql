-- +goose Up
-- +goose StatementBegin
ALTER TABLE tbl_checkout_payments ADD COLUMN payment_intent_id TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE tbl_checkout_payments DROP COLUMN payment_intent_id;
-- +goose StatementEnd

