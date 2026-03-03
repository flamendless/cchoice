-- +goose Up
-- +goose StatementBegin
ALTER TABLE tbl_staff_attendances RENAME COLUMN useragent_id TO in_useragent_id;
ALTER TABLE tbl_staff_attendances ADD COLUMN out_useragent_id INTEGER REFERENCES tbl_useragents(id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE tbl_staff_attendances DROP COLUMN out_useragent_id;
ALTER TABLE tbl_staff_attendances RENAME COLUMN in_useragent_id TO useragent_id;
-- +goose StatementEnd
