#!/usr/bin/env bash
set -euo pipefail

E2E_DB_NAME="${E2E_DB_NAME:-e2e.db}"
E2E_DB_PATH="file:./${E2E_DB_NAME}"
ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT_DIR"

# Password123 — same convention as customer/admin e2e fixtures.
E2E_STAFF_PASSWORD_HASH='$2a$10$bxxt5UWXU3QJXGMe1tM8cO4MVqkS5zU4eXxBI32ORXHgkDS7deWPq'

bootstrap_db() {
	echo "Bootstrapping e2e database..."
	rm -f "$E2E_DB_NAME" "${E2E_DB_NAME}-shm" "${E2E_DB_NAME}-wal"

	if [[ ! -x ./tmp/goose ]]; then
		echo "Building goose..."
		git submodule update --init --recursive ./cmd/goose
		(
			cd ./cmd/goose
			go mod tidy
			go build -tags="no_postgres,no_mysql,no_clickhouse,no_mssql,no_vertica,no_ydb" -o ../../tmp/goose ./cmd/goose
		)
		chmod +x ./tmp/goose
	fi

	go tool sqlc generate

	export DB_URL="$E2E_DB_PATH"
	export GOOSE_DBSTRING="$E2E_DB_PATH"
	./tmp/goose up

	go run -tags="fts5,staticfs" ./main.go parse_products \
		-p assets/xlsx/bosch.xlsx \
		-s DATABASE -t BOSCH \
		--use_db --db_path "$E2E_DB_PATH" \
		--verify_prices=1 --panic_on_error=1 \
		--images_basepath=./cmd/web/static/images/product_images/bosch/original/ \
		--images_format=webp

	# Seed migration marks slug scripts applied before products exist; run explicitly.
	go run -tags="fts5,staticfs" ./main.go populate_product_slugs --dry-run=false
}

seed_staff() {
	if [[ ! -f "$E2E_DB_NAME" ]]; then
		echo "e2e-seed: $E2E_DB_NAME not found; skipping staff seed"
		return
	fi

	echo "Seeding e2e staff accounts..."
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
}

if [[ -f "$E2E_DB_NAME" && "${E2E_FORCE_DB:-}" != "1" ]]; then
	echo "e2e.db exists; skipping bootstrap (set E2E_FORCE_DB=1 to rebuild)"
else
	bootstrap_db
fi

seed_staff

echo "e2e database ready: $E2E_DB_NAME"
