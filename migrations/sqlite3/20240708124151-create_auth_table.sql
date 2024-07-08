
-- +migrate Up
CREATE TABLE tbl_auth (
	id INTEGER PRIMARY KEY,
	user_id INTEGER NOT NULL,
	token TEXT NOT NULL,
	otp_enabled BOOLEAN NOT NULL DEFAULT false,
	otp_secret TEXT,
	recovery_codes TEXT,

	FOREIGN KEY (user_id) REFERENCES tbl_user(id)
);

CREATE INDEX idx_tbl_auth_otp_enabled ON tbl_auth(otp_enabled);

-- +migrate Down
DROP TABLE tbl_auth;
DROP INDEX idx_tbl_auth_otp_enabled;
