#!/usr/bin/env bash

# script for Linux (WSL-compatible) dev workflow
# @Brandon Blanker Lim-it

set -euf -o pipefail

DBNAME="test.db"
DBPATH="file:./${DBNAME}"

BROWSER="${BROWSER:-vivaldi}"
ISWSL=false
if grep -qi Microsoft /proc/version; then
	ISWSL=true
fi

serve() {
	local -; set -x;
	genall
	if "${ISWSL}"; then
		cmd.exe /c "start vivaldi http://localhost:7331/cchoice"
	fi
	go tool templ generate --watch --proxy="http://localhost:2626" --open-browser=false &
	go tool air -c ".air.api.toml" api
}

customrun() {
	go run ./main.go "${@:2}"
}

setup() {
	local -; set -x;
	if [ ! -f "./.git/hooks/pre-commit" ]; then
		cp "./scripts/pre-commit-unit-test.sh" "./.git/hooks/pre-commit"
		chmod +x "./.git/hooks/pre-commit"
	fi
}

clean() {
	local -; set -x;
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
	local -; set -x;
	clean
	gensql
	go tool github.com/rubenv/sql-migrate/sql-migrate up

	# local otherbrands=("BRADFORD" "SPARTAN" "SHINSETSU" "REDMAX" "KOBEWEL")
	# for brand in "${otherbrands[@]}"; do
	# 	go run ./main.go parse_xlsx -p "assets/xlsx/sample.xlsx" -t "${brand}" --use_db --db_path "${DBPATH}" --panic_on_error=1
	# done

	go run ./main.go thumbnailify_images --inpath="./cmd/web/static/images/product_images/bosch" --outpath="./cmd/web/static/thumbnails/product_images/bosch" --format="webp" --width=96 --height=96
	go run ./main.go thumbnailify_images --inpath="./cmd/web/static/images/product_images/bosch" --outpath="./cmd/web/static/images/product_images/bosch" --format="webp" --width=1080 --height=1080
	go run ./main.go convert_images --inpath="./cmd/web/static/images/brand_logos" --outpath="./cmd/web/static/images/brand_logos" --format="webp"

	# TODO: (Brandon) - there is a bug with update where the newly inserted tbl_products_categories.product_id are incorrect
	# go run ./main.go parse_xlsx -p "assets/xlsx/Price_List_effective_25_August_2023_r2.xlsx" -s "2023 PRICE LIST" -t "DELTAPLUS" --use_db --db_path "${DBPATH}" --verify_prices=1 --panic_on_error=1
	# go run ./main.go parse_xlsx -p "assets/xlsx/Price_List_effective_25_August_2023_r2.xlsx" -s "2023 PRICE LIST" -t "DELTAPLUS" --use_db --db_path "${DBPATH}" --verify_prices=1 --panic_on_error=1
	go run ./main.go parse_xlsx -p "assets/xlsx/bosch.xlsx" -s "DATABASE" -t "BOSCH" --use_db --db_path "${DBPATH}" --verify_prices=1 --panic_on_error=1 --images_basepath="./cmd/web/static/images/product_images/bosch/" --images_format="webp"
	# go run ./main.go parse_xlsx -p "assets/xlsx/bosch.xlsx" -s "DATABASE" -t "BOSCH" --use_db --db_path "${DBPATH}" --verify_prices=1 --panic_on_error=1 --images_basepath="./cmd/web/static/images/product_images/bosch/"

}

deps() {
	local -; set -x;
	# Tailwind https://tailwindcss.com/docs/installation/tailwind-cli
	curl -fsSL https://get.pnpm.io/install.sh | env PNPM_VERSION=10.0.0 sh -
	pnpm install tailwindcss @tailwindcss/cli

	# libvips https://www.libvips.org/install.html (I use Arch BTW)
	yay -S base-devel glib2 expat1 libdeflate libvips
}

gensql() {
	go tool sqlc generate
}

gentempl() {
	pnpx @tailwindcss/cli -m -i ./cmd/web/static/css/main.css -o ./cmd/web/static/css/tailwind.css
	go tool templ generate templ -v
}

genall() {
	go generate ./...
	gensql
	gentempl
}

sc() {
	local -; set -x;
	go fmt ./...
	go mod tidy
	go vet ./...
	go tool templ fmt ./cmd/web/components

	go tool usestdlibvars ./...
	go tool fatcontext ./...
	go tool betteralign -apply ./...
	go tool nilaway ./...
	go tool prealloc ./...
	go tool smrcptr ./...
	go tool unconvert ./...
	go run golang.org/x/tools/gopls/internal/analysis/modernize/cmd/modernize@latest -test ./...

	local PKGS
	PKGS=$(go list ./... | grep -v "internal/database/queries" | tr "\n" " ")
	for d in ${PKGS}; do
		go tool errcheck "$d"
	done

	set +f
	local GODIRS
	GODIRS=$(go list -f "{{.Dir}}" ./...)
	for d in ${GODIRS}; do
		if [[ ! $d == *"cmd/web/components"* ]]; then
			go tool goimports -w -local -v "$d"/*.go
		fi
	done
	set -f

	go tool govulncheck ./...
}

testall() {
	go tool gotestsum \
		--debug \
		--format=pkgname-and-test-fails \
		--format-icons=default \
		--format-hide-empty-pkg \
		--hide-summary=skipped \
		-- -cover -shuffle=on -race -test.v ./...
}

benchmark() {
	go test -bench=. -benchmem ./...
}

prod() {
	# genall
	# templ generate
	go build -o ./tmp/main .
	echo "Run: ./tmp/main api > out 2>&1 &"
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
	time "$1" "$@"
fi
