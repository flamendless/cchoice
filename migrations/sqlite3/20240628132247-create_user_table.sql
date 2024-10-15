
-- +migrate Up
CREATE TABLE tbl_user (
	id INTEGER PRIMARY KEY,
	first_name TEXT NOT NULL,
	middle_name TEXT NOT NULL,
	last_name TEXT NOT NULL,
	email TEXT NOT NULL,
	password TEXT NOT NULL,
	mobile_no TEXT NOT NULL,
	user_type TEXT CHECK (user_type IN ('API', 'SYSTEM')) NOT NULL DEFAULT 'API',
	status TEXT CHECK (status in ('ACTIVE', 'INACTIVE', 'DELETED')) NOT NULL DEFAULT 'ACTIVE',

	created_at DATETIME NOT NULL DEFAULT (DATE('1970-01-01 00:00:00')),
	updated_at DATETIME NOT NULL DEFAULT (DATE('1970-01-01 00:00:00')),
	deleted_at DATETIME NOT NULL DEFAULT (DATE('1970-01-01 00:00:00')),

	UNIQUE (email),
	UNIQUE (mobile_no)
);

CREATE INDEX idx_tbl_user_user_type ON tbl_user(user_type);
CREATE INDEX idx_tbl_user_status ON tbl_user(status);

-- +migrate Down
DROP INDEX idx_tbl_user_status;
DROP INDEX idx_tbl_user_user_type;
DROP TABLE tbl_user;
