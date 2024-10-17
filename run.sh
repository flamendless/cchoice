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
	genall
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
	gensql
	sql-migrate up

	local otherbrands=("BRADFORD" "SPARTAN" "SHINSETSU" "REDMAX" "KOBEWEL")
	for brand in "${otherbrands[@]}"; do
		go run ./main.go parse_xlsx -p "assets/xlsx/sample.xlsx" -t "${brand}" --use_db --db_path "${DBPATH}" --panic_on_error=1
	done

	# TODO: (Brandon) - there is a bug with update where the newly inserted tbl_products_categories.product_id are incorrect
	go run ./main.go parse_xlsx -p "assets/xlsx/Price_List_effective_25_August_2023_r2.xlsx" -s "2023 PRICE LIST" -t "DELTAPLUS" --use_db --db_path "${DBPATH}" --verify_prices=1 --panic_on_error=1
	# go run ./main.go parse_xlsx -p "assets/xlsx/Price_List_effective_25_August_2023_r2.xlsx" -s "2023 PRICE LIST" -t "DELTAPLUS" --use_db --db_path "${DBPATH}" --verify_prices=1 --panic_on_error=1
	go run ./main.go parse_xlsx -p "assets/xlsx/bosch.xlsx" -s "DATABASE" -t "BOSCH" --use_db --db_path "${DBPATH}" --verify_prices=1 --panic_on_error=1 --images_basepath="./client/static/images/product_images/bosch/"
	# go run ./main.go parse_xlsx -p "assets/xlsx/bosch.xlsx" -s "DATABASE" -t "BOSCH" --use_db --db_path "${DBPATH}" --verify_prices=1 --panic_on_error=1 --images_basepath="./client/static/images/product_images/bosch/"

}

deps() {
	go install github.com/rubenv/sql-migrate/...@latest
	go install github.com/air-verse/air@latest
	go install github.com/alexkohler/prealloc@latest
	go install github.com/nikolaydubina/smrcptr@latest
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	go install gotest.tools/gotestsum@latest
	go install go.uber.org/nilaway/cmd/nilaway@latest
	go install github.com/kisielk/errcheck@latest

	local VER="27.0"
	local PB_REL="https://github.com/protocolbuffers/protobuf/releases"
	curl -L "$PB_REL/download/v${VER}/protoc-${VER}-linux-x86_64.zip" -o "$HOME/protoc_${VER}.zip"
	unzip "$HOME/protoc_${VER}.zip" -d "$HOME/.local"
	export PATH="$PATH:$HOME/.local/bin"
}

gensql() {
	sqlc generate
}

genproto() {
	set +f
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/*.proto
	set -f
}

gentempl() {
	npx tailwindcss build -i client/static/css/style.css -o client/static/css/tailwind.css -m
	templ generate templ -v
}

genall() {
	go generate ./...
	gensql
	genproto
	gentempl
}

check() {
	go mod tidy
	templ fmt ./client/components

	set +f
	local gofiles=( internal/**/*.go conf/*.go grpc_server/**/*.go cmd/*.go client/*.go client/**/*.go )
	for file in "${gofiles[@]}"; do
		if [[ ! $file == *_templ.go ]]; then
			goimports -w -local -v "$file"
		fi
	done
	set -f

	go vet ./...
	prealloc ./...
	smrcptr ./...
	nilaway ./...
	errcheck ./...
}

testall() {
	gotestsum \
		--format=pkgname-and-test-fails \
		--format-icons=text \
		--format-hide-empty-pkg \
		--hide-summary=skipped \
		-- -cover -shuffle=on -race -test.v ./...
}

benchmark() {
	genall
	go test -bench=. -benchmem ./...
}

if [ "$#" -eq 0 ]; then
	echo "First use: chmod +x ${0}"
	echo "Usage: ${0}"
	echo "Commands:"
	echo "    benchmark"
	echo "    check"
	echo "    clean"
	echo "    cleandb"
	echo "    customrun"
	echo "    genall"
	echo "    genproto"
	echo "    gensql"
	echo "    gentempl"
	echo "    grpc"
	echo "    grpc_ui"
	echo "    setup"
	echo "    testall"
else
	echo "Running ${1}"
	"$1" "$@"
fi
