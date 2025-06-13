# Setup

- `go mod download`
- `go mod tidy`
- `chmod +x ./run.sh`
- `./run.sh deps`
- `./run.sh setup`
- `./run.sh genall`
- `./run.sh testall`
- `./run.sh benchmark`

# Env Vars

Create a `.env` file
```sh
# Required env vars
ADDRESS="localhost"
PORT=2626
PORT_FS=7070
APP_ENV="local" # "local" or "prod"
DB_URL=file:./test.db

# Optional env vars
# Only applicable if APP_ENV="local"
# 0 Debug
# 1 Info
# 2 Warn
# 3 Error
# 4 DPanic
# 5 Panic
# 6 Fatal
LOG_MIN_LEVEL=1
USESSL=0
USEHTTP2=1

# Only applicable if USESSL=1
CERTPATH="/home/<user>/certs/localhost+2.pem"
KEYPATH="/home/<user>/certs/localhost+2-key.pem"
```

# Generate local cert

```
mkdir ~/certs
cd ~/certs
mkcert -install
mkcert localhost 127.0.0.1 ::1
```

# Commits
- Maintenance:
    - simple fix
    - simple revision
    - code quality
- Toolings:
    - go tools
    - dev tools
    - library upgrade
    - dependency upgrade
- Script:
    - run.sh
    - cmd/thumbnailify_images.go
- Config:
    - git-chglog
    - dotenv
    - air
- Docs:
    - README
    - Changelogs
- Performance:
    - optimization
- Server
- Web
