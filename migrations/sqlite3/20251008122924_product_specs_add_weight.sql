-- +goose Up
ALTER TABLE tbl_product_specs ADD COLUMN weight_unit VARCHAR(2);
ALTER TABLE tbl_product_specs ADD COLUMN weight FLOAT;

-- +goose Down
ALTER TABLE tbl_product_specs DROP COLUMN weight;
ALTER TABLE tbl_product_specs DROP COLUMN weight_unit;
