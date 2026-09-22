#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT_DIR/e2e"

pnpm install --frozen-lockfile

if [[ "${1:-}" == "--ui" ]]; then
	pnpm run test:ui
else
	pnpm test
fi
