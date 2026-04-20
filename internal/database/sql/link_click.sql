-- name: CreateLinkClick :exec
INSERT INTO tbl_link_clicks (
    link_id,
    clicked_at,
    referrer,
    user_agent,
    ip_hash,
    device,
    utm_source,
    utm_medium,
    utm_campaign
) VALUES (
    ?,
    datetime('now'),
    ?, ?, ?, ?, ?, ?, ?
);

-- name: GetLinkClicksByLinkID :many
SELECT
    id,
    link_id,
    clicked_at,
    referrer,
    user_agent,
    ip_hash,
    device,
    utm_source,
    utm_medium,
    utm_campaign
FROM tbl_link_clicks
WHERE link_id = ?
ORDER BY clicked_at DESC;

-- name: CountLinkClicksByLinkID :one
SELECT COUNT(*) as count
FROM tbl_link_clicks
WHERE link_id = ?;
