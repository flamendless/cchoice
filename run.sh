#!/usr/bin/env bash

# script for Linux (WSL-compatible) dev workflow
# @Brandon Blanker Lim-it

set -euf -o pipefail

DBNAME="test.db"

clean() {
	if [ -f "./${DBNAME}" ]; then
		rm "${DBNAME}"
	fi
	if [ -f "./${DBNAME}-shm" ]; then
		rm "${DBNAME}-shm"
	fi
	if [ -f "./${DBNAME}-wal" ]; then
		rm "${DBNAME}-wal"
	fi

	sqlc generate
	sql-migrate up
}

testall() {
	clean
	GOOS=linux go run ./main.go parse_xlsx -p "xlsx/Price_List_effective_25_August_2023_r2.xlsx" -s "2023 PRICE LIST" -t "delta_plus" --use_db --db_path "file:./test.db" --verify_prices=1
	GOOS=linux go run ./main.go parse_xlsx -p "xlsx/Price_List_effective_25_August_2023_r2.xlsx" -s "2023 PRICE LIST" -t "delta_plus" --use_db --db_path "file:./test.db" --verify_prices=1
	GOOS=linux go run ./main.go parse_xlsx -p "xlsx/bosch.xlsx" -s "DATABASE" -t "bosch" --use_db --db_path "file:./test.db" --verify_prices=1
	GOOS=linux go run ./main.go parse_xlsx -p "xlsx/bosch.xlsx" -s "DATABASE" -t "bosch" --use_db --db_path "file:./test.db" --verify_prices=1
}

if [ "$#" -eq 0 ]; then
	echo "First use: chmod +x ${0}"
	echo "Usage: ${0} clean | testall"
else
	"$1" "$@"
fi
