#!/usr/bin/env bash

# script for Linux (WSL-compatible) dev workflow
# @Brandon Blanker Lim-it

set -euf -o pipefail
source .env

GOOS="linux"
DBNAME="test.db"
DBPATH="file:./${DBNAME}"

GRPC_SERVER_ADDR=":50051"
CLIENTPORT="3001"

WIN_PATH=/mnt/c/Windows/System32
alias cmd.exe="$WIN_PATH"/cmd.exe

grpc_ui() {
	cmd.exe /c "start vivaldi http://127.0.0.1:36477/"
	local token=$(go run ./main.go jwt -s "issue" -a "API" -o "true" -u "client@cchoice.com")
	local auth="authorization: bearer ${token}"
	grpcui \
		-authority "bearer" \
		-reflect-header "${auth}" \
		-rpc-header "${auth}" \
		-port 36477 \
		-plaintext "${GRPC_SERVER_ADDR}"
}

grpc() {
	air -c ".air.grpc.toml" serve_grpc \
		-r=1 --db_path "${DBPATH}" \
		--address "${GRPC_SERVER_ADDR}" \
		--log_payload_received=true
}

client() {
	cmd.exe /c "start vivaldi http://localhost:7331/"
	templ generate --watch --proxy="http://localhost:${CLIENTPORT}" --open-browser=false &
	air -c ".air.client.toml" serve_client \
		-p ":${CLIENTPORT}" --grpc_address "${GRPC_SERVER_ADDR}"
}

customrun() {
	go run ./main.go "${@:2}"
}

setup() {
	if [ ! -f "./.git/hooks/pre-commit" ]; then
		cp "./scripts/pre-commit-unit-test.sh" "./.git/hooks/pre-commit"
		chmod +x "./.git/hooks/pre-commit"
	fi
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

cleandb() {
	clean
	genall
	gensql

	local otherbrands=("BRADFORD" "SPARTAN" "SHINSETSU" "REDMAX" "KOBEWEL")
	for brand in "${otherbrands[@]}"; do
		go run ./main.go parse_xlsx -p "assets/xlsx/sample.xlsx" -t "${brand}" --use_db --db_path "${DBPATH}" --panic_on_error=1
	done

	go run ./main.go parse_xlsx -p "assets/xlsx/Price_List_effective_25_August_2023_r2.xlsx" -s "2023 PRICE LIST" -t "DELTAPLUS" --use_db --db_path "${DBPATH}" --verify_prices=1 --panic_on_error=1
	go run ./main.go parse_xlsx -p "assets/xlsx/Price_List_effective_25_August_2023_r2.xlsx" -s "2023 PRICE LIST" -t "DELTAPLUS" --use_db --db_path "${DBPATH}" --verify_prices=1 --panic_on_error=1
	go run ./main.go parse_xlsx -p "assets/xlsx/bosch.xlsx" -s "DATABASE" -t "BOSCH" --use_db --db_path "${DBPATH}" --verify_prices=1 --panic_on_error=1 --images_basepath="./client/static/images/product_images/bosch/"
	go run ./main.go parse_xlsx -p "assets/xlsx/bosch.xlsx" -s "DATABASE" -t "BOSCH" --use_db --db_path "${DBPATH}" --verify_prices=1 --panic_on_error=1 --images_basepath="./client/static/images/product_images/bosch/"

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
	echo "running gentempl..."
	npx tailwindcss build -i client/static/css/style.css -o client/static/css/tailwind.css -m
	templ generate templ -v
}

templlr() {
	echo "running templlr..."
	templ generate --notify-proxy -v
}

genall() {
	echo "running genall..."
	go generate ./...
	gensql
	genproto
	gentempl
}

check() {
	echo "running check..."
	go mod tidy

	set +f
	local gofiles=( internal/**/*.go conf/*.go grpc_server/**/*.go cmd/*.go )
	for file in "${gofiles[@]}"; do
		goimports -w -local -v "$file"
	done
	set -f

	go vet ./...
	prealloc ./...
	smrcptr ./...
}

testall() {
	genall
	go test ./...
}

benchmark() {
	go test -bench=. -benchmem ./...
}

if [ "$#" -eq 0 ]; then
	echo "First use: chmod +x ${0}"
	echo "Usage: ${0}"
	echo "Commands:"
	echo "    setup"
	echo "    check"
	echo "    clean"
	echo "    genall"
	echo "    genproto"
	echo "    gensql"
	echo "    grpc"
	echo "    grpc_ui"
	echo "    gentempl"
	echo "    cleandb"
	echo "    customrun"
	echo "    testall"
	echo "    benchmark"
else
	"$1" "$@"
fi
