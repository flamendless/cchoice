-- +goose Up
ALTER TABLE tbl_product_specs ADD COLUMN weight_unit VARCHAR(2) CHECK (weight_unit IN ('kg', 'g', 'lb', 'oz') OR weight_unit IS NULL);
ALTER TABLE tbl_product_specs ADD COLUMN weight FLOAT;

-- +goose Down
ALTER TABLE tbl_product_specs DROP COLUMN weight;
ALTER TABLE tbl_product_specs DROP COLUMN weight_unit;
