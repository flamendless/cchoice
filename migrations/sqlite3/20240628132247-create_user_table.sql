
-- +migrate Up
CREATE TABLE tbl_user (
	id INTEGER PRIMARY KEY,
	first_name TEXT NOT NULL,
	middle_name TEXT NOT NULL,
	last_name TEXT NOT NULL,
	email TEXT NOT NULL,
	mobile_no TEXT NOT NULL,
	user_type TEXT CHECK (user_type IN ('USER', 'SYSTEM', 'ADMIN')) NOT NULL DEFAULT 'USER',

	created_at DATETIME NOT NULL DEFAULT (DATE('1970-01-01 00:00:00')),
	updated_at DATETIME NOT NULL DEFAULT (DATE('1970-01-01 00:00:00')),
	deleted_at DATETIME NOT NULL DEFAULT (DATE('1970-01-01 00:00:00')),

	UNIQUE (email),
	UNIQUE (mobile_no)
);

CREATE INDEX idx_tbl_user_user_type ON tbl_user(user_type);

-- +migrate Down
DROP TABLE tbl_user;
DROP INDEX idx_tbl_user_user_type;
