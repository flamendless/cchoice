-- +goose Up
-- +goose StatementBegin
INSERT INTO tbl_settings(name, value)
VALUES
	('url_waze', 'https://www.waze.com/en/live-map/directions/ph/calabarzon/general-trias/c-choice-construction-supply?place=ChIJ0wcjerfVlzMRG16aIHSShxU')
;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM tbl_settings
WHERE name IN ('url_waze');
-- +goose StatementEnd
