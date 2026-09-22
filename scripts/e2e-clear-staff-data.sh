#!/usr/bin/env bash
set -euo pipefail

# Clear mutable e2e staff data from e2e.db.
# Usage: e2e-clear-staff-data.sh [attendance|time-off|all]
# Default: all

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT_DIR"

E2E_DB_NAME="${E2E_DB_NAME:-e2e.db}"
E2E_STAFF_EMAIL="${E2E_STAFF_EMAIL:-e2e-staff@cchoice.test}"
SCOPE="${1:-all}"

if [[ ! -f "$E2E_DB_NAME" ]]; then
	echo "e2e-clear-staff-data: $E2E_DB_NAME not found; skipping"
	exit 0
fi

case "$SCOPE" in
	attendance | time-off | all) ;;
	*)
		echo "e2e-clear-staff-data: unknown scope '$SCOPE' (use attendance, time-off, or all)" >&2
		exit 1
		;;
esac

sqlite3 "$E2E_DB_NAME" <<SQL
DELETE FROM tbl_staff_attendances
WHERE staff_id = (
    SELECT id
    FROM tbl_staffs
    WHERE email = '${E2E_STAFF_EMAIL}'
        AND deleted_at = '1970-01-01 00:00:00+00:00'
)
AND '${SCOPE}' IN ('attendance', 'all');

DELETE FROM tbl_staff_time_offs
WHERE staff_id = (
    SELECT id
    FROM tbl_staffs
    WHERE email = '${E2E_STAFF_EMAIL}'
        AND deleted_at = '1970-01-01 00:00:00+00:00'
)
AND '${SCOPE}' IN ('time-off', 'all');
SQL
