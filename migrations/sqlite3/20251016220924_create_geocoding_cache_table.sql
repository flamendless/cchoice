
-- +goose Up
CREATE TABLE tbl_geocoding_cache (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	address TEXT NOT NULL,
	normalized_address TEXT NOT NULL,
	latitude TEXT NOT NULL,
	longitude TEXT NOT NULL,
	formatted_address TEXT NOT NULL,
	place_id TEXT,
	response_data TEXT,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	expires_at TIMESTAMP,

	UNIQUE (normalized_address)
);

CREATE INDEX idx_geocoding_cache_address ON tbl_geocoding_cache(address);
CREATE INDEX idx_geocoding_cache_normalized_address ON tbl_geocoding_cache(normalized_address);
CREATE INDEX idx_geocoding_cache_expires_at ON tbl_geocoding_cache(expires_at);

-- +goose Down
DROP INDEX idx_geocoding_cache_expires_at;
DROP INDEX idx_geocoding_cache_normalized_address;
DROP INDEX idx_geocoding_cache_address;
DROP TABLE tbl_geocoding_cache;

