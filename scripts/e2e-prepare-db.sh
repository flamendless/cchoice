#!/usr/bin/env bash
set -euo pipefail

E2E_DB_NAME="e2e.db"
E2E_DB_PATH="file:./${E2E_DB_NAME}"
ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT_DIR"

if [[ -f "$E2E_DB_NAME" && "${E2E_FORCE_DB:-}" != "1" ]]; then
	echo "e2e.db exists; skipping prepare (set E2E_FORCE_DB=1 to rebuild)"
	exit 0
fi

echo "Preparing e2e database..."
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

echo "e2e database ready: $E2E_DB_NAME"
