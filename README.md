# Setup

- `git clone --recurse-submodules --shallow-submodules -j8 <repo url>`
- `cd cchoice`
- `go mod download`
- `go mod tidy`
- `go install tool`
- `go install github.com/magefile/mage@latest`
- `mage deps`
- `mage setup`
    - If you get `cp: cannot create regular file './git/hooks/pre-commit': No such file or directory
    Error: exit status 1` open `magefile.go` find line 280 and change `./git/hooks/pre-commit` to `./.git/hooks/pre-commit` and run     `mage setup` again.
    - **Important**: After running `mage setup`, you MUST edit `.env` and fill in the `BUSINESS_*` variables. The application will 
    panic on startup if these are empty.
- `mage setupprod`
- `mage genall`
- `mage cleandb`
- `mage genimages`
- `mage genmaps`
- `mage testall`
- `mage benchmark`
    
- Add "Env Vars" Details
Update the current brief reference:
# Env Vars
See `.env.sample`
**Mandatory Fields (must be filled in `.env`):**
- `BUSINESS_LAT`, `BUSINESS_LNG`, `BUSINESS_ADDRESS`, `BUSINESS_LINE1`, `BUSINESS_LINE2`, `BUSINESS_CITY`, `BUSINESS_STATE`, `BUSINESS_POSTAL_CODE`, `BUSINESS_COUNTRY`

# Running
- NOTE: if you don't have `vivaldi` browser, open `magefile.go` find this line `browser = getenv("BROWSER", "vivaldi")` 
change `vivaldi` to your preferred browser (e.g., `browser = getenv("BROWSER", "chrome")`).   
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
