#!/usr/bin/env bash

# script for Linux (WSL-compatible) dev workflow
# @Brandon Blanker Lim-it

set -euf -o pipefail

GOOS="linux"
DBNAME="test.db"
WIN_PATH=/mnt/c/Windows/System32
alias cmd.exe="$WIN_PATH"/cmd.exe

http() {
	go run ./main.go serve_http
}

grpc() {
	go run ./main.go serve_grpc
}

grpc_ui() {
	cmd.exe /c "start vivaldi http://127.0.0.1:36477/"
	grpcui -port 36477 -plaintext ":50051"
}

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

	gensql
}

testall() {
	clean
	go run ./main.go parse_xlsx -p "xlsx/Price_List_effective_25_August_2023_r2.xlsx" -s "2023 PRICE LIST" -t "delta_plus" --use_db --db_path "file:./test.db" --verify_prices=1
	go run ./main.go parse_xlsx -p "xlsx/Price_List_effective_25_August_2023_r2.xlsx" -s "2023 PRICE LIST" -t "delta_plus" --use_db --db_path "file:./test.db" --verify_prices=1
	go run ./main.go parse_xlsx -p "xlsx/bosch.xlsx" -s "DATABASE" -t "bosch" --use_db --db_path "file:./test.db" --verify_prices=1
	go run ./main.go parse_xlsx -p "xlsx/bosch.xlsx" -s "DATABASE" -t "bosch" --use_db --db_path "file:./test.db" --verify_prices=1
}

deps() {
	VER="27.0"
	PB_REL="https://github.com/protocolbuffers/protobuf/releases"
	curl -L "$PB_REL/download/v${VER}/protoc-${VER}-linux-x86_64.zip" -o "$HOME/protoc_${VER}.zip"
	unzip "$HOME/protoc_${VER}.zip" -d "$HOME/.local"
	export PATH="$PATH:$HOME/.local/bin"
}

gensql() {
	sqlc generate
	sql-migrate up
	check
}

genproto() {
	set +f
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/*.proto
	set -f
	check
}

check() {
	goimports -w -local -v .
	go vet ./...
	prealloc ./...
	smrcptr ./...
}

if [ "$#" -eq 0 ]; then
	echo "First use: chmod +x ${0}"
	echo "Usage: ${0}"
	echo "Commands:"
	echo "    clean"
	echo "    check"
	echo "    testall"
	echo "    http"
	echo "    grpc"
	echo "    genproto"
	echo "    gensql"
else
	"$1" "$@"
fi
