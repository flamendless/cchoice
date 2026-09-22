#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT_DIR"

E2E_DB_NAME="${E2E_DB_NAME:-e2e.db}"

if [[ ! -f "$E2E_DB_NAME" ]]; then
	echo "e2e-clear-staff-attendance: $E2E_DB_NAME not found; skipping"
	exit 0
fi

sqlite3 "$E2E_DB_NAME" <<'SQL'
DELETE FROM tbl_staff_attendances
WHERE staff_id = (
    SELECT id
    FROM tbl_staffs
    WHERE email = 'e2e-staff@cchoice.test'
        AND deleted_at = '1970-01-01 00:00:00+00:00'
);
SQL
