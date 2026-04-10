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

# NOTES:

For GH workflow, comment out `[ -z "$PS1" ] && return` in server's .bashrc
