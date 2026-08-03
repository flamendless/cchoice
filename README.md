# Setup

- `git clone --recurse-submodules --shallow-submodules -j8 <repo url>`
- `cd cchoice`
- `go mod download`
- `go mod tidy`
- `go install tool`
- `go install github.com/magefile/mage@latest`
- `mage deps`
- `mage setup`
- `mage setupprod`
- `mage genall`
- `mage cleandb`
- `mage dbup`
- `mage genimages`
- `mage genmaps`
- `mage testall`
- `mage benchmark`

# Env Vars

See `.env.sample`

# Running

Users should set their own `BROWSER` env var in their shell. Example: `export BROWSER=chrome`
- run `mage serve`
- or `mage serveweb` for faster iteration for frontend changes only

---

# Generate local cert (for STAGING and PROD environment only)

```
mkdir ~/certs
cd ~/certs
mkcert -install
mkcert localhost 127.0.0.1 ::1
```

---

# Generate HMAC Secret (for C-Points token signing)

```bash
openssl rand -base64 32
```
Set in `.env`:
```
CPOINT_HMAC_SECRET="your-generated-secret"
```

---

# Commit Prefix
- Feature
- Maintenance:
    - simple fix or revision
    - code quality
- Deps:
    - library/dep upgrade
- Toolings:
    - go or dev tools
- Script:
    - run.sh, scripts/*, cmd/*, magefile.go
- CICD:
    - gh actions, workflows
- Config:
    - git-chglog
    - dotenv
    - air
- Docs:
    - README
    - Changelogs
- Performance
- Server:
    - SQL
    - Migrations
    - API
- Web
    - E-Mail
- Bugfix
- AI

---

# VERSION:

- To check current version in browser. Invoke `await _G.VERSION()` in browser console

---

# METRICS:

- In `.env`, provide values for `BASIC_AUTH_USERNAME` and `BASIC_AUTH_PASSWORD_HASH`
- Use `gen_basic_auth` to hash value for `BASIC_AUTH_PASSWORD_HASH`
- You can now access `/metrics`

---

# NOTES:

- For GH workflow, comment out `[ -z "$PS1" ] && return` in server's .bashrc.

---

# Post-migrate scripts

After `mage dbup`, some schema changes require one-off CLI backfills. These are registered in `internal/datamigrate/manifest.go` and tracked in `tbl_datamigrate_applied`. Successful runs (with `--dry-run=false`) are recorded automatically; `mage serve` and prod/dev builds fail until required scripts are applied.

Check status (lists pending goose migrations first, then post-migrate scripts):

```bash
mage datamigratestatus
# or after mage build:
./tmp/main datamigrate status
```

Registered scripts (run in goose migration order — `AfterVersion` ascending):

| # | Script | After migration | Command |
|---|--------|-----------------|---------|
| 1 | `apply_discount:sale_2025` | `20260109164336` (2026-01-09) | `./tmp/main apply_discount -i scripts/csv/sale_2025.csv --dry-run=false` |
| 2 | `populate_product_images_cdn` | `20260327151912`; only when `STORAGE_PROVIDER` is Cloudflare Images | `./tmp/main populate_product_images_cdn --dry-run=false` |
| 3 | `populate_product_slugs` | `20260418065629` | `./tmp/main populate_product_slugs --dry-run=false` |
| 4 | `populate_brand_slugs` | `20260710120000` | `./tmp/main populate_brand_slugs --dry-run=false` |

Already-applied scripts on existing databases are seeded by goose migration `20260803140000_seed_tbl_datamigrate_applied.sql` after you run `mage dbup`.

Break-glass manual mark (use only when a script was applied outside the normal CLI path):

```bash
./tmp/main datamigrate mark <script-name>
```
