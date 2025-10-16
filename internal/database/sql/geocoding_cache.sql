-- name: GetGeocodingCacheByAddress :one
SELECT
	id, address, normalized_address, latitude, longitude,
	formatted_address, place_id, response_data,
	created_at, updated_at, expires_at
FROM tbl_geocoding_cache
WHERE normalized_address = ?
	AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP)
LIMIT 1;

-- name: InsertGeocodingCache :one
INSERT INTO tbl_geocoding_cache (
	address,
	normalized_address,
	latitude,
	longitude,
	formatted_address,
	place_id,
	response_data,
	expires_at
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
RETURNING id, address, normalized_address, latitude, longitude,
	formatted_address, place_id, response_data,
	created_at, updated_at, expires_at;

-- name: UpdateGeocodingCache :exec
UPDATE tbl_geocoding_cache
SET
	address = ?,
	latitude = ?,
	longitude = ?,
	formatted_address = ?,
	place_id = ?,
	response_data = ?,
	updated_at = CURRENT_TIMESTAMP,
	expires_at = ?
WHERE normalized_address = ?;

-- name: UpsertGeocodingCache :one
INSERT INTO tbl_geocoding_cache (
	address,
	normalized_address,
	latitude,
	longitude,
	formatted_address,
	place_id,
	response_data,
	expires_at
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(normalized_address) DO UPDATE SET
	address = excluded.address,
	latitude = excluded.latitude,
	longitude = excluded.longitude,
	formatted_address = excluded.formatted_address,
	place_id = excluded.place_id,
	response_data = excluded.response_data,
	updated_at = CURRENT_TIMESTAMP,
	expires_at = excluded.expires_at
RETURNING id, address, normalized_address, latitude, longitude,
	formatted_address, place_id, response_data,
	created_at, updated_at, expires_at;

-- name: DeleteExpiredGeocodingCache :exec
DELETE FROM tbl_geocoding_cache
WHERE expires_at IS NOT NULL
	AND expires_at <= CURRENT_TIMESTAMP;

-- name: GetGeocodingCacheStats :one
SELECT
	COUNT(*) as total_entries,
	COUNT(CASE WHEN expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP THEN 1 END) as active_entries,
	COUNT(CASE WHEN expires_at IS NOT NULL AND expires_at <= CURRENT_TIMESTAMP THEN 1 END) as expired_entries
FROM tbl_geocoding_cache;

