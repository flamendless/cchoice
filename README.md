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
ADDRESS="localhost"
PORT=8080
PORT_FS=7070
APP_ENV="local"
DB_URL=file:./test.db
USEHTTP2=1
USESSL=1

#only if APP_ENV="local" and USESSL=1
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
