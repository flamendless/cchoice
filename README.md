# Setup

- `git clone --recurse-submodules --shallow-submodules -j8 <repo>`
- `go mod download`
- `go mod tidy`
- `chmod +x ./run.sh`
- `./run.sh deps`
- `./run.sh setup`
- `./run.sh genall`
- `./run.sh cleandb`
- `./run.sh genimages`
- `./run.sh testall`
- `./run.sh benchmark`

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
    - run.sh, scripts/*, cmd/*
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

---

# NOTES:

For GH workflow, comment out `[ -z "$PS1" ] && return` in server's .bashrc
