# Setup

- `go mod download`
- `go mod tidy`
- `chmod +x ./run.sh`
- `./run.sh deps`
- `./run.sh setup`
- `./run.sh genall`
- `./run.sh testall`
- `./run.sh benchmark`

# Env Varas

Create a `.env` file
```sh
ADDRESS="localhost"
PORT=8080
APP_ENV=local
DB_URL=file:./test.db
USEHTTP2=1
```
