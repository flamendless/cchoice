# AGENTS.md - Developer Guide for cchoice

This file provides guidelines and commands for agentic coding agents working in this repository.

---

## Build Commands

### Setup (First Time)
```bash
git clone --recurse-submodules --shallow-submodules -j8 <repo url>
cd cchoice
go mod download
go mod tidy
go install tool
go install github.com/magefile/mage@latest
mage deps
mage setup
mage setupprod
mage genall
```

### Development
```bash
mage serve       # Run API + Web with hot reload (port 7331)
mage serveweb   # Run Web only with hot reload
mage serveadmin # Open admin panel
mage servecustomer # Open customer panel
mage build     # Build main binary to ./tmp/main
mage buildweb  # Build web only
```

### Code Generation
```bash
mage genall     # Generate all (sqlc + templ + version)
mage gensql     # Generate SQL (sqlc)
mage gentempl  # Generate templ components
mage genversion # Generate version file
mage genimages # Generate/convert product images
```

### Database
```bash
mage dbup       # Run pending migrations
mage dbdown    # Rollback last migration
mage cleandb   # Clean DB and re-parse products
```

### Migration
Use this command to create new migration files:
```bash
./tmp/goose create <filename> sql # Create new migration
```

- always set defaults of created_at and updated_at to use `datetime('now')`
- always set created_at and updated_at in create/update queries

NEVER RUN/APPLY MIGRATIONS. USERS SHOULD MANUALLY DO THAT.

ALWAYS USE `go tool` to run commands like:
- go tool templ
- go tool air
- go tool sqlc

ENUMS:
- always use go stringer for enums
- always do `strings.ToUpper` the enum string in the `switch` block

---

## Lint Commands

```bash
mage scf        # Run all: fmt, vet, templ fmt, betteralign, nilaway, unconvert, modernize, govulncheck
```

---

## Test Commands

### Run All Tests
```bash
mage testall    # Unit tests + trailing whitespace check + lint
go test ./...   # Run all unit tests
```

### Run Single Test
```bash
go test ./internal/services -run TestGenerateCode
go test ./path/to/package -run TestName
go test ./internal/services/cpoint_test.go -run TestGenerateCode_Uniqueness
```

### Integration Tests
```bash
mage testinteg   # Run integration tests for changed packages
mage testSum     # Run tests with gotestsum (shuffle, race detection)
mage benchmark  # Run benchmarks
```

### Specific Package Tests
```bash
go test ./internal/services/... -v
go test ./internal/utils/... -run TestValidate
go test ./internal/payments/... -bench=.
```

## Building and Generation
- Always use `mage build` to build the project.
- Always use `mage gentempl` to generate templ files.
- Always use `mage gensql` to generate sql files.


---

## Code Style Guidelines

### Import Organization
Imports must be organized in three groups (use `go fmt` automatically):
1. Standard library (`"database/sql"`, `"context"`, etc.)
2. External packages (`"github.com/stretchr/testify"`, `"github.com/go-chi/chi/v5"`, etc.)
3. Internal packages (`"cchoice/internal/..."`)

```go
import (
    "context"
    "strings"

    "github.com/stretchr/testify/assert"
    "github.com/go-chi/chi/v5"

    "cchoice/internal/database"
    "cchoice/internal/errs"
)
```

### Naming Conventions
- **Packages**: lowercase, short (`services`, `utils`, `httputil`)
- **Interfaces**: `I` prefix for interface names (`IEncode`, `IService`)
- **Structs**: PascalCase (`CustomerService`, `RegisterCustomerParams`)
- **Fields**: PascalCase (`FirstName`, `CustomerType`)
- **Functions**: PascalCase for exported, camelCase for unexported
- **Constants**: PascalCase for exported, camelCase for unexported
- **Database field params**: `queries.CreateCustomerParams` style from sqlc

### File Layout
- Declare all `type`, `struct`, and `interface` definitions at the top of each file, before `const`, `var`, and functions
- When a package grows large, split by responsibility into multiple files in the same package (e.g. types, constants, parsing, merge logic) instead of one long file
- Keep related tests in similarly named files (e.g. `product_export_parse_test.go` for `product_export_parse.go`)

### Error Handling
- Return errors directly: `return nil, err`
- Wrap errors with context: `return 0, fmt.Errorf("failed to create customer: %w", err)`
- Use sentinel errors: `var ErrInvalidCode = errors.New("invalid code")` in `internal/errs`
- Check errors explicitly: `if err != nil { return ... }`
- Convention: Avoid things like `fmt.Errorf("invalid user type")`, return `errs.ErrInvalidUserType` instead
- if the only return value is `error`, then always check it within if statement. Sample:
```go
//DO NOT DO THIS
err := foo()
if err != nil {...}

//DO THIS INSTEAD
if err := foo(); err != nil {...}
```
- if there are multiple return values but the others are discarded except the `err`, do this as well:
```go
//DO NOT DO THIS
_, err := foo()
if err != nil {...}

//DO THIS INSTEAD
if _, err := foo(); err != nil {...}
- if error is in http handlers, use redirectHX with utils.URLWithError

### Types
- Use `sql.NullString`, `sql.NullInt64` for nullable DB fields
- Use `null.String` from `github.com/goccy/go-json` for JSON fields
- Use `time.Time` for timestamps
- Use `int64` for IDs
- Use `context.Context` as first parameter for service methods

### Build Tags
Use these tags when building/running:
- `fts5` - Enable FTS5 search
- `staticfs` - Embed static files
- `imageprocessing` - Enable image processing

```bash
go build -tags="fts5,staticfs,imageprocessing" -o ./tmp/main
go run -tags="fts5,staticfs" ./main.go
```

### Database (sqlc)
- SQL queries in `internal/database/queries/query.sql`
- Generated code in `internal/database/queries/`
- Use `mage gensql` to regenerate
- Parameter structs: `queries.CreateCustomerParams{...}`
- Always use UPPERCASE for text values like in CHECK CONSTRAINTS
- Always use UPPERCASE for default values for TEXT like `status TEXT NOT NULL DEFAULT 'DRAFT',`

### Templates (templ)
- Templates in `cmd/web/components/`
- Use `mage gentempl` to generate Go code
- Format with `go tool templ fmt ./cmd/web/components`
- always use enums if possible instead of string

### Testing
- Use `github.com/stretchr/testify/assert` for assertions
- Use table-driven tests with anonymous structs
- Use `t.Run(name, func(t *testing.T))` for sub-tests
- Benchmark with `BenchmarkXxx(b *testing.B)`

### Logging
- Use `go.uber.org/zap` for structured logging
- Global logger via `logs.Log().Info` or `logs.Log().Warn`

### Interfaces
- Whenever an implementor will satisfy/implement an interface, do this as an example `var _ IService = (*QRService)(nil)` trick at the bottom for compile-time assurance

### IDs
- Always pass database ID as encoded strings (to functions, frontend/templ, etc) and then decode it at the service level
- Always check decoded ID with like `if decoded == encode.INVALID { ... }`

### HTMX
- Use `redirectHX` and combinations of utils.URL or utils.URLWithError or utils.URLWithSuccess
- Use `utils.URL` always in templ src or hx-

### Links / URLs
- always build urls with utils.URL like `utils.URL("/test")` or `utils.URL(fmt.Sprintf(...))`

### Time
- Always define/use date/time layout in constants package instead of hardcoded strings
- always do above for date/time parsing
- always do above for date/time format

### Email templates
- create templates in ./templates
- always reference existing templates like `templates/customer_verification.html`
- always add cchoice logo in header

### Any
- use `any` instead of `interface{}`

### Handlers
- In handlers, define `const logtag = "[FOO]"` for logging
- In handlers that have redirects, use `const page = "url"`

### Admin Services
- In services that have CRUD like holidays creation, products creation, and so on, record the activity in `StaffLogsService`.
- Follow convention in `reports.go`:
```go
	result := "success"
	defer func() {
		if err := s.staffLog.CreateLog(ctx, adminStaffID, "export", "attendance_report_xlsx", result, nil); err != nil {
			logs.LogCtx(ctx).Error("[ReportService] failed to log xlsx report generation", zap.Error(err))
		}
	}()

    if err := foo(); err != nil {
        result = err.Error()
    }
```

### Queries
- instead of passing updated at values from go -> sql, define the updated_at = NOW in sqlc code

### Services
- always accept and return ids as strings and do decode/encode in the service

### More
- Use recommendations by `modernize` tools with go 1.26 and above as basis
- always define regex in constants/regex.go, use regexp.MustCompile
- avoid passing created_at, updated_at, deleted_at if possible IF it can be set in the sql level

---

## Project Structure

```
cchoice/
├── cmd/              # CLI commands
│   ├── web/         # Web components (templ)
│   ├── parse_products/
│   └── parse_map/
├── internal/       # Internal packages
│   ├── services/   # Business logic

│   ├── database/  # DB service and queries
│   ├── utils/     # Utility functions
│   ├── httputil/  # HTTP utilities
│   ├── errs/      # Sentinel errors
│   ├── enums/     # Enumerations
│   └── conf/      # Configuration
├── tmp/            # Temporary build output
├── scripts/        # Utility scripts
├── templates/      # HTML templates
└── magefile.go     # Build tasks
```

---

## Commit Message Prefixes

- `Feature` - New features
- `Maintenance` - Simple fixes, code quality
- `Deps` - Library updates
- `Toolings` - Dev tools
- `Script` - Shell scripts
- `CICD` - GH actions
- `Config` - Configuration
- `Docs` - Documentation
- `Bugfix` - Bug fixes

---

## Environment

- Go 1.26+
- SQLite database
- chi router
- templ for HTML
- sqlc for DB
- air for hot reload

For `.env` setup, copy `.env.sample` to `.env` and fill in values.

---

DO NOT RUN `mage sc` and `mage scf` and `mage dbup` and `mage dbdown`

---

## Cursor Cloud specific instructions

Dependencies (Go tools, `mage`, `goose`, tailwind) are refreshed automatically by the startup update script; system libs (`libvips-dev`, `libmagickwand-dev`, `openslide-tools`, `libxml2-dev`, `libjxl-dev`, `libsqlite3-dev`, `pkg-config`) are baked into the VM snapshot. `mage` lives at `$(go env GOPATH)/bin` (`~/go/bin`), which is on `PATH` in login shells.

### Running the storefront without external secrets (WEB mode)
- Run in **WEB mode** to exercise the storefront/admin UI without PayMongo/Lalamove/Google Maps/Maileroo credentials. In `LOCAL`/`api` mode those services are `env-required` and the server **panics on boot** (e.g. `[GMAPS]: API key is required`).
- WEB mode is selected by `APP_ENV` in `.env`. Note: `.air.web.toml` sets `APP_ENV=web`, but `godotenv/autoload` reading `.env` wins, so the value in `.env` is authoritative — set `APP_ENV="web"` there.
- `.env` also requires (validated even with `USESSL=0`): non-empty `CERTPATH`/`KEYPATH` (dummy paths are fine when `USESSL=0` since certs load only when SSL is on) and **all** `BUSINESS_*` fields non-empty, else boot panics.
- The app auto-loads `.env` (via `godotenv/autoload`); `goose` also auto-loads `.env`.

### Reliable run vs. hot reload
- `mage serveweb` starts `templ generate --watch` + tailwind watch + air. On startup the templ watcher rewrites all `*_templ.go` files and **races** air's first build, causing transient `expected 'package', found 'EOF'` failures (large files like `tracked_links_templ.go`). It sometimes recovers on the next file change.
- For a reliable run, build and run the binary directly: `go build -tags="fts5 staticfs" -o ./tmp/web . && APP_ENV=web ./tmp/web web`. The storefront serves on **http://localhost:2626/cchoice** (port 2626; the `7331` proxy only exists under the templ watcher).
- The app panics with `open sink "logs/app.log"` if `./logs` is missing. `mage setup` has an early-return that skips creating `logs/` when it also had to create `.env` on the same run — the update script does `mkdir -p logs` to cover this.

### Database & seed data
- SQLite `test.db`. Apply migrations with the goose binary: `GOOSE_DRIVER=sqlite3 GOOSE_DBSTRING="file:./test.db" GOOSE_MIGRATION_DIR="./migrations/sqlite3" ./tmp/goose up` (per repo policy, migrations are applied manually, never via `mage dbup`).
- Seed products from the bundled spreadsheet: `go run -tags="fts5 staticfs" ./main.go parse_products -p assets/xlsx/bosch.xlsx -s DATABASE -t BOSCH --use_db --db_path file:./test.db --verify_prices=1 --panic_on_error=1 --images_basepath=./cmd/web/static/images/product_images/bosch/original/ --images_format=webp` (or `mage cleandb`, which also resets the DB).
- Seed caveats (data, not environment): `parse_products` inserts products with **empty `slug`**, so `/product/{slug}` links 404 (slugs are only generated when products are created via the admin `ProductService`). Product images are referenced as `.webp` but the repo ships `.png` originals, so product/cart thumbnails are blank until `mage genimages` converts them.

### Routes
- The storefront product catalog is at `/cchoice/` (products lazy-load via HTMX from `/cchoice/product-categories/...`). There is no `/cchoice/shop` route.
