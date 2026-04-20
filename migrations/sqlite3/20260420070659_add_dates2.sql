-- +goose Up
-- +goose StatementBegin
ALTER TABLE tbl_products_categories ADD COLUMN created_at DATETIME;
ALTER TABLE tbl_products_categories ADD COLUMN updated_at DATETIME;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
