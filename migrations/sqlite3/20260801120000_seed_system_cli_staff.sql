-- +goose Up
-- +goose StatementBegin
-- System CLI staff used as the actor for CLI command audit logs (tbl_staff_logs).
-- Override attribution with CCHOICE_CLI_STAFF_ID env var (encoded staff ID).
INSERT INTO tbl_staffs (
    first_name,
    middle_name,
    last_name,
    birthdate,
    sex,
    date_hired,
    time_in_schedule,
    time_out_schedule,
    position,
    user_type,
    email,
    mobile_no,
    password,
    require_in_shop,
    status,
    created_at,
    updated_at,
    deleted_at
)
SELECT
    'System',
    NULL,
    'CLI',
    '1970-01-01',
    'M',
    '1970-01-01',
    NULL,
    NULL,
    'System CLI',
    'SUPERUSER',
    'cli@system.internal',
    '00000000000',
    '$2a$10$nHyQUwg1TV0cdGTTXpDVvOWzx3DRUwtbg3Z/k4tvqRNeB/HjIKbeG',
    0,
    'REGULAR',
    datetime('now'),
    datetime('now'),
    '1970-01-01 00:00:00+00:00'
WHERE NOT EXISTS (
    SELECT 1
    FROM tbl_staffs
    WHERE email = 'cli@system.internal'
        AND deleted_at = '1970-01-01 00:00:00+00:00'
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
UPDATE tbl_staffs
SET
    deleted_at = datetime('now'),
    updated_at = datetime('now')
WHERE email = 'cli@system.internal'
    AND deleted_at = '1970-01-01 00:00:00+00:00';
-- +goose StatementEnd
