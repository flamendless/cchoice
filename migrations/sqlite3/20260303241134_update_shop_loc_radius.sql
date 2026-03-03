-- +goose Up
-- +goose StatementBegin
UPDATE tbl_settings SET value = '{"lat": 14.333199079659577, "lng": 120.88151883134833, "radius_meters": 60}' WHERE name = 'shop_location'
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
UPDATE tbl_settings SET value = '{"lat": 14.333199079659577, "lng": 120.88151883134833, "radius_meters": 26}' WHERE name = 'shop_location'
-- +goose StatementEnd
