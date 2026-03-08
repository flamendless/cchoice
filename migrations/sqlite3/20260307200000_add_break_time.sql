-- +goose Up
-- +goose StatementBegin
ALTER TABLE tbl_staff_attendances ADD COLUMN lunch_break_in TEXT;
ALTER TABLE tbl_staff_attendances ADD COLUMN lunch_break_in_location TEXT;
ALTER TABLE tbl_staff_attendances ADD COLUMN lunch_break_in_useragent_id INTEGER REFERENCES tbl_useragents(id);
ALTER TABLE tbl_staff_attendances ADD COLUMN lunch_break_out TEXT;
ALTER TABLE tbl_staff_attendances ADD COLUMN lunch_break_out_location TEXT;
ALTER TABLE tbl_staff_attendances ADD COLUMN lunch_break_out_useragent_id INTEGER REFERENCES tbl_useragents(id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE tbl_staff_attendances DROP COLUMN lunch_break_in;
ALTER TABLE tbl_staff_attendances DROP COLUMN lunch_break_in_location;
ALTER TABLE tbl_staff_attendances DROP COLUMN lunch_break_in_useragent_id;
ALTER TABLE tbl_staff_attendances DROP COLUMN lunch_break_out;
ALTER TABLE tbl_staff_attendances DROP COLUMN lunch_break_out_location;
ALTER TABLE tbl_staff_attendances DROP COLUMN lunch_break_out_useragent_id;
-- +goose StatementEnd
