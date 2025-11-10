-- +goose Up
ALTER TABLE tbl_brand_images ADD COLUMN s3_url TEXT;

-- +goose Down
ALTER TABLE tbl_brand_images DROP COLUMN s3_url;

