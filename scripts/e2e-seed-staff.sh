#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT_DIR"

E2E_DB_NAME="${E2E_DB_NAME:-e2e.db}"

if [[ ! -f "$E2E_DB_NAME" ]]; then
	echo "e2e-seed-staff: $E2E_DB_NAME not found; skipping"
	exit 0
fi

# Password123 — same convention as customer e2e fixtures.
E2E_STAFF_PASSWORD_HASH='$2a$10$bxxt5UWXU3QJXGMe1tM8cO4MVqkS5zU4eXxBI32ORXHgkDS7deWPq'

sqlite3 "$E2E_DB_NAME" <<SQL
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
    'Test',
    'E2E',
    'Superuser',
    '1990-01-15',
    'M',
    '2020-01-01',
    NULL,
    NULL,
    'E2E Superuser',
    'SUPERUSER',
    'e2e-superuser@cchoice.test',
    '9171234567',
    '${E2E_STAFF_PASSWORD_HASH}',
    0,
    'REGULAR',
    datetime('now'),
    datetime('now'),
    '1970-01-01 00:00:00+00:00'
WHERE NOT EXISTS (
    SELECT 1
    FROM tbl_staffs
    WHERE email = 'e2e-superuser@cchoice.test'
        AND deleted_at = '1970-01-01 00:00:00+00:00'
);

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
    'Test',
    'E2E',
    'Staff',
    '1990-01-15',
    'M',
    '2020-01-01',
    NULL,
    NULL,
    'E2E Staff',
    'STAFF',
    'e2e-staff@cchoice.test',
    '9171234568',
    '${E2E_STAFF_PASSWORD_HASH}',
    0,
    'REGULAR',
    datetime('now'),
    datetime('now'),
    '1970-01-01 00:00:00+00:00'
WHERE NOT EXISTS (
    SELECT 1
    FROM tbl_staffs
    WHERE email = 'e2e-staff@cchoice.test'
        AND deleted_at = '1970-01-01 00:00:00+00:00'
);
SQL
