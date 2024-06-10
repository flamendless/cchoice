#!/usr/bin/env bash

# script for Linux (WSL-compatible) dev workflow
# @Brandon Blanker Lim-it

set -euf -o pipefail

GOOS="linux"
DBNAME="test.db"
WIN_PATH=/mnt/c/Windows/System32
alias cmd.exe="$WIN_PATH"/cmd.exe

PRE_CMD="./run.sh check && ./run.sh genall"

http() {
	air --build.bin "./tmp/main serve_http" --build.pre_cmd "${PRE_CMD}"
}

grpc() {
	air --build.bin "./tmp/main serve_grpc -r=1" --build.pre_cmd "${PRE_CMD}"
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
	echo "running gensql..."
	echo "running sqlc..."
	sqlc generate
	echo "running sql-migrate..."
	sql-migrate up
}

genproto() {
	echo "running genproto..."
	set +f
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/*.proto
	set -f
}

genall() {
	echo "running genall..."
	gensql
	genproto
}

check() {
	echo "running check..."
	goimports -w -local -v .
	go vet ./...
	prealloc ./...
	smrcptr ./...
}

if [ "$#" -eq 0 ]; then
	echo "First use: chmod +x ${0}"
	echo "Usage: ${0}"
	echo "Commands:"
	echo "    check"
	echo "    clean"
	echo "    genall"
	echo "    genproto"
	echo "    gensql"
	echo "    grpc"
	echo "    grpc_ui"
	echo "    http"
	echo "    testall"
else
	"$1" "$@"
fi
