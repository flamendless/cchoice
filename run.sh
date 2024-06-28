#!/usr/bin/env bash

# script for Linux (WSL-compatible) dev workflow
# @Brandon Blanker Lim-it

set -euf -o pipefail
source .env

GOOS="linux"
DBNAME="test.db"
DBPATH="file:./${DBNAME}"

GRPC_SERVER_ADDR=":50051"

WIN_PATH=/mnt/c/Windows/System32
alias cmd.exe="$WIN_PATH"/cmd.exe

grpc() {
	air serve_grpc -r=1 --db_path "${DBPATH}" --address "${GRPC_SERVER_ADDR}" --log_payload_received=true
}

grpc_ui() {
	cmd.exe /c "start vivaldi http://127.0.0.1:36477/"
	grpcui -authority "bearer" -reflect-header "authorization: bearer grpcui" -port 36477 -plaintext "${GRPC_SERVER_ADDR}"
}

client() {
	local clientport="3001"
	cmd.exe /c "start vivaldi http://localhost:${clientport}/"
	air serve_client -p ":${clientport}" --grpc_address "${GRPC_SERVER_ADDR}"
}

customrun() {
	go run ./main.go "${@:2}"
}

clean() {
	echo "cleaning ${DBNAME}..."
	if [ -f "./${DBNAME}" ]; then
		rm "${DBNAME}"
	fi
	if [ -f "./${DBNAME}-shm" ]; then
		rm "${DBNAME}-shm"
	fi
	if [ -f "./${DBNAME}-wal" ]; then
		rm "${DBNAME}-wal"
	fi
}

testall() {
	clean
	gensql
	go run ./main.go parse_xlsx -p "assets/xlsx/Price_List_effective_25_August_2023_r2.xlsx" -s "2023 PRICE LIST" -t "delta_plus" --use_db --db_path "${DBPATH}" --verify_prices=1 --panic_on_error=1
	go run ./main.go parse_xlsx -p "assets/xlsx/Price_List_effective_25_August_2023_r2.xlsx" -s "2023 PRICE LIST" -t "delta_plus" --use_db --db_path "${DBPATH}" --verify_prices=1 --panic_on_error=1
	go run ./main.go parse_xlsx -p "assets/xlsx/bosch.xlsx" -s "DATABASE" -t "bosch" --use_db --db_path "${DBPATH}" --verify_prices=1 --panic_on_error=1
	go run ./main.go parse_xlsx -p "assets/xlsx/bosch.xlsx" -s "DATABASE" -t "bosch" --use_db --db_path "${DBPATH}" --verify_prices=1 --panic_on_error=1
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

gentempl() {
	templ generate templ
}

genall() {
	echo "running genall..."
	go generate
	npx tailwindcss build -i client/static/css/style.css -o client/static/css/tailwind.css -m
	gensql
	genproto
	gentempl
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
	echo "    gentempl"
	echo "    testall"
	echo "    customrun"
else
	"$1" "$@"
fi
