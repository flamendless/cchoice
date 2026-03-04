-- +goose Up
-- +goose StatementBegin
ALTER TABLE tbl_staff_attendances RENAME COLUMN location TO in_location;
ALTER TABLE tbl_staff_attendances ADD COLUMN out_location TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE tbl_staff_attendances DROP COLUMN out_location;
ALTER TABLE tbl_staff_attendances RENAME COLUMN in_location TO location;
-- +goose StatementEnd
