# Setup

- `git clone --recurse-submodules --shallow-submodules -j8 <repo url>`
- `go mod download`
- `go mod tidy`
- `go install github.com/magefile/mage@latest`
- `mage deps`
- `mage setup`
- `mage setupprod`
- `mage genall`
- `mage cleandb`
- `mage genimages`
- `mage genmaps`
- `mage testall`
- `mage benchmark`

# Env Vars

See `.env.sample`

# Running

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

---

# VERSION:

- To check current version in browser. Invoke `await _G.VERSION()` in browser console

---

# NOTES:

For GH workflow, comment out `[ -z "$PS1" ] && return` in server's .bashrc
