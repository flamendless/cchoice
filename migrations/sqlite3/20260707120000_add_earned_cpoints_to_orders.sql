-- +goose Up
ALTER TABLE tbl_orders ADD COLUMN earned_cpoints INTEGER NOT NULL DEFAULT 0;

-- +goose Down
ALTER TABLE tbl_orders DROP COLUMN earned_cpoints;
