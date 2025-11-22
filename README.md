# Setup

- `git clone --recurse-submodules --shallow-submodules -j8 <repo>`
- `go mod download`
- `go mod tidy`
- `go install github.com/magefile/mage@latest`
- `mage deps`
- `mage setup`
- `mage setupprod`
- `mage genAll`
- `mage cleanDB`
- `mage genImages`
- `mage genMaps`
- `mage testAll`
- `mage benchmark`

# Env Vars

See `.env.sample`

# Generate local cert

```
mkdir ~/certs
cd ~/certs
mkcert -install
mkcert localhost 127.0.0.1 ::1
```

# Commits
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
- Bugfix

---

# VERSION:

- To check current version in browser. Invoke `await _G.VERSION()` in browser console

---

# NOTES:

For GH workflow, comment out `[ -z "$PS1" ] && return` in server's .bashrc
