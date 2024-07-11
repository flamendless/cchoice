
-- +migrate Up
CREATE TABLE tbl_auth (
	id INTEGER PRIMARY KEY,
	user_id INTEGER NOT NULL,
	token TEXT NOT NULL,
	otp_enabled BOOLEAN NOT NULL DEFAULT false,
	otp_secret TEXT,
	recovery_codes TEXT,
	otp_status TEXT CHECK (
		otp_status IN ('INITIAL', 'ENROLLED', 'SENT_CODE', 'VALID')
	) NOT NULL DEFAULT 'INITIAL',

	FOREIGN KEY (user_id) REFERENCES tbl_user(id)
);

CREATE INDEX idx_tbl_auth_otp_enabled ON tbl_auth(otp_enabled);
CREATE INDEX idx_tbl_auth_otp_status ON tbl_auth(otp_status);

-- +migrate Down
DROP TABLE tbl_auth;
DROP INDEX idx_tbl_auth_otp_enabled;
