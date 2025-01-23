#!/usr/bin/env bash

# script for Linux (WSL-compatible) dev workflow
# @Brandon Blanker Lim-it

set -euf -o pipefail

GOOS="linux"
DBNAME="test.db"
DBPATH="file:./${DBNAME}"

BROWSER="${BROWSER:-vivaldi}"
ISWSL=false
if [[ $(grep -i Microsoft /proc/version) ]]; then
	ISWSL=true
fi

serve() {
	genall
	if "${ISWSL}"; then
		cmd.exe /c "start vivaldi http://localhost:7331/"
	fi
	templ generate --watch --proxy="http://localhost:8080" --open-browser=false &
	air -c ".air.api.toml" api
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

	# local otherbrands=("BRADFORD" "SPARTAN" "SHINSETSU" "REDMAX" "KOBEWEL")
	# for brand in "${otherbrands[@]}"; do
	# 	go run ./main.go parse_xlsx -p "assets/xlsx/sample.xlsx" -t "${brand}" --use_db --db_path "${DBPATH}" --panic_on_error=1
	# done

	# TODO: (Brandon) - there is a bug with update where the newly inserted tbl_products_categories.product_id are incorrect
	# go run ./main.go parse_xlsx -p "assets/xlsx/Price_List_effective_25_August_2023_r2.xlsx" -s "2023 PRICE LIST" -t "DELTAPLUS" --use_db --db_path "${DBPATH}" --verify_prices=1 --panic_on_error=1
	# go run ./main.go parse_xlsx -p "assets/xlsx/Price_List_effective_25_August_2023_r2.xlsx" -s "2023 PRICE LIST" -t "DELTAPLUS" --use_db --db_path "${DBPATH}" --verify_prices=1 --panic_on_error=1
	go run ./main.go parse_xlsx -p "assets/xlsx/bosch.xlsx" -s "DATABASE" -t "BOSCH" --use_db --db_path "${DBPATH}" --verify_prices=1 --panic_on_error=1 --images_basepath="./cmd/web/static/images/product_images/bosch/"
	# go run ./main.go parse_xlsx -p "assets/xlsx/bosch.xlsx" -s "DATABASE" -t "BOSCH" --use_db --db_path "${DBPATH}" --verify_prices=1 --panic_on_error=1 --images_basepath="./cmd/web/static/images/product_images/bosch/"

}

deps() {
	go install github.com/a-h/templ/cmd/templ@latest
	go install github.com/rubenv/sql-migrate/...@latest
	go install github.com/air-verse/air@latest
	go install github.com/alexkohler/prealloc@latest
	go install github.com/nikolaydubina/smrcptr@latest
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	go install gotest.tools/gotestsum@latest
	go install go.uber.org/nilaway/cmd/nilaway@latest
	go install github.com/kisielk/errcheck@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest
	go install github.com/dkorunic/betteralign/cmd/betteralign@latest
}

gensql() {
	sqlc generate
}

gentempl() {
	pnpx @tailwindcss/cli -i ./cmd/web/static/css/main.css -o ./cmd/web/static/css/output.css
	templ generate templ -v
}

genall() {
	go generate ./...
	gensql
	gentempl
}

sc() {
	go mod tidy
	templ fmt ./cmd/web/components

	go vet ./...
	prealloc ./...
	smrcptr ./...
	nilaway ./...
	errcheck ./...
	govulncheck ./...
	betteralign -apply ./...

	set +f
	local gofiles=( internal/**/*.go conf/*.go cmd/*.go cmd/**/*.go )
	for file in "${gofiles[@]}"; do
		if [[ ! $file == *_templ.go ]]; then
			goimports -w -local -v "$file"
		fi
	done
	set -f
}

testall() {
	gotestsum \
		--format=pkgname-and-test-fails \
		--format-icons=default \
		--format-hide-empty-pkg \
		--hide-summary=skipped \
		-- -cover -shuffle=on -race -test.v ./...
}

benchmark() {
	go test -bench=. -benchmem ./...
}

if [ "$#" -eq 0 ]; then
	echo "First use: chmod +x ${0}"
	echo "Usage: ${0}"
	echo "Commands:"
	echo "    benchmark"
	echo "    clean"
	echo "    cleandb"
	echo "    customrun"
	echo "    genall"
	echo "    gensql"
	echo "    gentempl"
	echo "    serve"
	echo "    sc"
	echo "    setup"
	echo "    testall"
else
	echo "Running ${1}"
	"$1" "$@"
fi
