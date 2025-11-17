
-- +goose Up
INSERT INTO tbl_settings(name, value)
VALUES
	('url_tiktok', 'https://www.tiktok.com/@cchoicesales?_t=8pPsHyIgtF4&_r=1'),
	('url_facebook', 'https://www.facebook.com/profile.php?id=61553625688578&mibextid=ZbWKwL'),
	('url_gmap', 'https://maps.app.goo.gl/JZCZfbseZuh7eYZg7'), --TODO: (Brandon)
	('email', 'cchoicesales23@gmail.com'),
	('mobile_no', '09976894824'),
	('address', 'EMC Katiwala Bldg. Bagong Kalsada, Pasong Kawayan 2, General Trias, Cavite')
;

-- +goose Down
DELETE FROM tbl_settings
WHERE name IN ('url_tiktok', 'url_facebook', 'url_gmap', 'email', 'mobile_no', 'address')
;
