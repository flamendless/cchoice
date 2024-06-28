
-- +migrate Up
INSERT INTO tbl_user(
	first_name,
	middle_name,
	last_name,
	email,
	mobile_no,
	user_type
) VALUES
	(
		'grpcui',
		'grpcui',
		'grpcui',
		'grpcui@cchoice.com',
		'+639000000000',
		'SYSTEM'
	),
	(
		'client',
		'client',
		'client',
		'client@cchoice.com',
		'+639000000001',
		'API'
	)
;

-- +migrate Down
DELETE FROM tbl_user
WHERE user_type = 'SYSTEM' AND email = 'grpcui@cchoice.com';

DELETE FROM tbl_user
WHERE user_type = 'API' AND email = 'client@cchoice.com';
