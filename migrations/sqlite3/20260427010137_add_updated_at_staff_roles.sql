-- +goose Up
-- +goose StatementBegin
ALTER TABLE tbl_staff_roles ADD COLUMN updated_at DATETIME;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE tbl_staff_roles DROP COLUMN updated_at;
-- +goose StatementEnd
