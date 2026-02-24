-- +goose Up
-- +goose StatementBegin
INSERT INTO tbl_settings(name, value)
VALUES
	('shop_location', '{"lat": 14.333199079659577, "lng": 120.88151883134833, "radius_meters": 26}')
;

ALTER TABLE tbl_staff_attendances ADD COLUMN location TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM tbl_settings
WHERE name = 'shop_location';
-- +goose StatementEnd
