#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT_DIR"

E2E_PORT="${E2E_PORT:-7332}"
export APP_ENV=LOCAL
export DB_URL="file:./e2e.db"
export GOOSE_DBSTRING="file:./e2e.db"
export PORT="$E2E_PORT"
export FSMODE="${FSMODE:-staticfs}"

if [[ ! -f ./e2e.db ]]; then
	bash scripts/e2e-prepare-db.sh
fi

if [[ ! -x ./tmp/main ]]; then
	bash scripts/e2e-build.sh
fi

bash scripts/e2e-seed-staff.sh

exec ./tmp/main api
