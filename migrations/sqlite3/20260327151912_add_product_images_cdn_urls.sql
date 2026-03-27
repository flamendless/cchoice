-- +goose Up
-- +goose StatementBegin
ALTER TABLE tbl_product_images ADD COLUMN cdn_url TEXT DEFAULT '';
ALTER TABLE tbl_product_images ADD COLUMN cdn_url_thumbnail TEXT DEFAULT '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE tbl_product_images DROP COLUMN cdn_url;
ALTER TABLE tbl_product_images DROP COLUMN cdn_url_thumbnail;
-- +goose StatementEnd
