#!/usr/bin/env bash

# script for Linux (WSL-compatible) dev workflow
# @Brandon Blanker Lim-it

set -euf -o pipefail

DBNAME="test.db"
DBPATH="file:./${DBNAME}"
TMP="./tmp"

BROWSER="${BROWSER:-vivaldi}"
ISWSL=false
ISMAC=false

if [[ $(uname) == "Darwin" ]]; then
	ISMAC=true
elif grep -qi Microsoft /proc/version; then
	ISWSL=true
fi

serve() {
	local -; set -x;
	genall

	if "${ISWSL}"; then
		cmd.exe /c "start ${BROWSER} http://localhost:7331/cchoice"
	elif "${ISMAC}"; then
		open -a ${BROWSER} "http://localhost:7331/cchoice"
	fi

	go tool templ generate --watch --proxy="http://localhost:2626" --open-browser=false &
	go tool air -c ".air.api.toml" api
}

build() {
	local -; set -x;
	genall
	# go build -tags='fts5' -tags="embeddedfs" -o "${TMP}/main" .
	go build -tags='fts5' -tags="staticfs" -o "${TMP}/main" .
}

buildgoose() {
	local -; set -x;
	git submodule update --init --recursive
	cd ./cmd/goose
	go mod tidy
	go build -tags='no_postgres no_mysql no_clickhouse no_mssql no_vertica no_ydb' -o "../../${TMP}/goose" ./cmd/goose
	cd ../..
	chmod +x "${TMP}/goose"
}

customrun() {
	FSMODE="stubfs" go run ./main.go "${@:2}"
}

setup() {
	local -; set -x;
	if [ ! -f "./.git/hooks/pre-commit" ]; then
		cp "./scripts/pre-commit-unit-test.sh" "./.git/hooks/pre-commit"
		chmod +x "./.git/hooks/pre-commit"
	fi

	if [ ! -f "./.env" ]; then
		cp "./.env.sample" "./.env"
	fi
}

genimages() {
	go run -tags=imageprocessing ./main.go thumbnailify_images --inpath="./cmd/web/static/images/product_images/bosch" --outpath="./cmd/web/static/thumbnails/product_images/bosch" --format="webp" --width=96 --height=96
	go run -tags=imageprocessing ./main.go thumbnailify_images --inpath="./cmd/web/static/images/product_images/bosch" --outpath="./cmd/web/static/thumbnails/product_images/bosch" --format="webp" --width=1080 --height=1080
	go run -tags=imageprocessing ./main.go convert_images --inpath="./cmd/web/static/images/brand_logos" --outpath="./cmd/web/static/images/brand_logos" --format="webp"

}

genmaps() {
	go run -tags="staticfs" ./main.go parse_map --filepath="./assets/xlsx/PSGC-2Q-2025-Publication-Datafile.xlsx" --json="true"
}

cleandb() {
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
	gensql
	"${TMP}/goose" up

	go run -tags="fts5" -tags="staticfs" ./main.go parse_products -p "assets/xlsx/bosch.xlsx" -s "DATABASE" -t "BOSCH" --use_db --db_path "${DBPATH}" --verify_prices=1 --panic_on_error=1 --images_basepath="./cmd/web/static/images/product_images/bosch/" --images_format="webp"

}

deps_arch() {
	echo "Installing dependencies for Arch..."
	yay -S --noconfirm \
		base-devel \
		glib2 \
		expat1 \
		libdeflate \
		libvips \
		libmagick \
		openslide \
		libxml2 \
		libjxl \
		golangci-lint-bin
}

deps_debian() {
	echo "Installing dependencies for Debian..."
	sudo apt update
	sudo apt install -y \
		build-essential \
		golang-go \
		git \
		sqlite3 \
		libsqlite3-dev \
		libvips-dev \
		libmagickwand-dev \
		openslide-tools \
		libxml2-dev \
		libjxl-dev \
		curl
}

deps_mac() {
	echo "Installing dependencies for MacOS..."
	brew install \
		go \
		git \
		sqlite \
		vips \
		imagemagick \
		openslide \
		libxml2 \
		jpeg-xl \
		curl \
		golangci-lint
}

deps() {
	local -; set -x;

	# Tailwind https://tailwindcss.com/docs/installation/tailwind-cli
	if [[ ! -f tailwindcss ]]; then
		local BIN
		if "${ISWSL}"; then
			BIN="tailwindcss-linux-x64"
		elif "${ISMAC}"; then
			BIN="tailwindcss-macos-arm64"
		else
			BIN="tailwindcss-linux-x64"
		fi

		curl -LO "https://github.com/tailwindlabs/tailwindcss/releases/latest/download/${BIN}"
		chmod +x "${BIN}"
		mv "${BIN}" tailwindcss
	else
		echo "You already have tailwindcss binary"
	fi

	if "${ISWSL}"; then
		local DISTROID
		DISTROID=$(grep ^ID= /etc/os-release | cut -d= -f2 | tr -d '"')

		if [[ "$DISTROID" == "arch" ]]; then
			deps_arch
		elif [[ "$DISTROID" == "debian" ]]; then
			deps_debian
		else
			echo "Unknown or unsupported distribution: $DISTROID"
			exit 1
		fi
	elif "${ISMAC}"; then
		deps_mac
	fi

	buildgoose
}

gensql() {
	go tool sqlc generate
}

gentempl() {
	./tailwindcss -m -i ./cmd/web/static/css/main.css -o ./cmd/web/static/css/tailwind.css
	go tool templ generate templ -v
}

genall() {
	go generate ./...
	gensql
	gentempl
}

genchlog() {
	go tool git-chglog -o CHANGELOGS.md
}

sc() {
	local -; set -x;
	go fmt ./...
	go mod tidy
	go vet ./...
	go tool templ fmt ./cmd/web/components

	go tool betteralign -apply ./...
	go tool nilaway ./...
	go tool smrcptr ./...
	go tool unconvert ./...
	go run golang.org/x/tools/gopls/internal/analysis/modernize/cmd/modernize@latest -test ./...

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
	if [ -x "$(command -v golangci-lint)" ]; then
		golangci-lint config verify
		golangci-lint run
	fi
	go test ./... -failfast "${@:2}"
}

testsum() {
	go tool gotestsum \
		--debug \
		--format=pkgname-and-test-fails \
		--format-icons=default \
		--format-hide-empty-pkg \
		--hide-summary=skipped \
		-- -cover -shuffle=on -race -test.v ./...
}

benchmark() {
	go test -bench=. -benchmem ./... "${@:2}"
}

prof() {
	if [ "$#" -ne 3 ]; then
		echo "must pass either 2 more arguments"
		exit 1
	fi
	MODE="benchmark" go test -cpuprofile "${TMP}/cpu.prof" -memprofile "${TMP}/mem.prof" -benchmem -bench=. -o "${TMP}/" "./${2}"
	go tool pprof -http=localhost:3031 "${TMP}/${3}.prof"
}

prod() {
	# genall
	# templ generate
	build
	echo "Run: ./tmp/main api > out 2>&1 &"
}

db() {
	"${TMP}/goose" "${@:2}"
}

if [ "$#" -eq 0 ]; then
	echo "First use: chmod +x ${0}"
	echo "Usage: ${0}"
	echo "Commands:"
	echo "    benchmark"
	echo "    cleandb"
	echo "    customrun"
	echo "    db"
	echo "    deps"
	echo "    genall"
	echo "    genchlog"
	echo "    genimages"
	echo "    genmaps"
	echo "    gensql"
	echo "    gentempl"
	echo "    prof"
	echo "    sc"
	echo "    serve"
	echo "    setup"
	echo "    testall"
	echo "    testsum"
else
	echo "Running ${1}"
	time "$1" "$@"
fi
