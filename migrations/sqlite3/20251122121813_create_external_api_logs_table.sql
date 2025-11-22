-- +goose Up
-- +goose StatementBegin
CREATE TABLE tbl_external_api_logs (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	checkout_id INTEGER,
	service TEXT NOT NULL,
	api TEXT NOT NULL,
	endpoint TEXT NOT NULL,
	http_method TEXT NOT NULL,
	payload TEXT,
	response TEXT,
	status_code INTEGER,
	error_message TEXT,
	is_successful INTEGER NOT NULL DEFAULT 0,
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

	FOREIGN KEY (checkout_id) REFERENCES tbl_checkouts(id)
);

CREATE INDEX idx_external_api_logs_checkout_id ON tbl_external_api_logs(checkout_id);
CREATE INDEX idx_external_api_logs_service ON tbl_external_api_logs(service);
CREATE INDEX idx_external_api_logs_api ON tbl_external_api_logs(api);
CREATE INDEX idx_external_api_logs_created_at ON tbl_external_api_logs(created_at);
CREATE INDEX idx_external_api_logs_is_successful ON tbl_external_api_logs(is_successful);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX idx_external_api_logs_is_successful;
DROP INDEX idx_external_api_logs_created_at;
DROP INDEX idx_external_api_logs_api;
DROP INDEX idx_external_api_logs_service;
DROP INDEX idx_external_api_logs_checkout_id;
DROP TABLE tbl_external_api_logs;
-- +goose StatementEnd

