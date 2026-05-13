#!/usr/bin/env bash
set -euo pipefail

usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -a, --app-dir <DIR>     CChoice app directory"
    echo "                          Default: script parent directory, or APP_DIR"
    echo "  -d, --db <NAME>         Database shortcut: 'prod' or 'test'"
    echo "                          Default: test when DB_URL is unset"
    echo "  -u, --db-url <URL>      SQLite DB URL/path"
    echo "                          Default: DB_URL from env/.env, or file:./test.db"
    echo "  -l, --local [DIR]       Save backup locally instead of object storage"
    echo "                          Default: LOCAL_BACKUP_DIR or ~/backups/cchoice"
    echo "  -r, --remote <REMOTE>   rclone destination"
    echo "                          Default: BACKUP_REMOTE or cchoice_backup:cchoice-assets/backups"
    echo "  -h, --help              Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0                                      # Backup DB to object storage"
    echo "  $0 -d prod                             # Backup prod.db to object storage"
    echo "  $0 --local                             # Backup DB to ~/backups/cchoice"
    echo "  $0 --local /path/to/backups            # Backup DB to a custom local directory"
    echo "  $0 --db-url file:/data/cchoice.db      # Backup a specific SQLite DB"
    exit 0
}

trim_env_value() {
    printf "%s" "$1" \
        | sed -e 's/[[:space:]]*#.*$//' \
              -e 's/^[[:space:]]*//' \
              -e 's/[[:space:]]*$//' \
              -e 's/^"//' \
              -e 's/"$//' \
              -e "s/^'//" \
              -e "s/'$//"
}

load_db_url_from_env_file() {
    local env_file="$1"
    local db_line

    [[ -f "$env_file" ]] || return 0

    db_line="$(grep -E '^[[:space:]]*DB_URL=' "$env_file" | tail -n 1 || true)"
    [[ -n "$db_line" ]] || return 0

    DB_URL="$(trim_env_value "${db_line#*=}")"
}

sqlite_db_path_from_url() {
    local db_url="$1"
    local path

    case "$db_url" in
        :memory:|file::memory:*)
            return 1
            ;;
        file:*)
            path="${db_url#file:}"
            ;;
        *)
            path="$db_url"
            ;;
    esac

    path="${path%%\?*}"

    # Normalize file:///absolute/path.db without pretending to support file://host/path.db.
    if [[ "$path" == ///* ]]; then
        path="${path#//}"
    elif [[ "$path" == //* ]]; then
        return 1
    fi

    if [[ "$path" != /* ]]; then
        path="$APP_DIR/$path"
    fi

    printf "%s" "$path"
}

require_command() {
    if ! command -v "$1" >/dev/null 2>&1; then
        echo "Error: required command not found: $1" >&2
        exit 1
    fi
}

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
APP_DIR="${APP_DIR:-$(cd "$SCRIPT_DIR/.." && pwd)}"
DB_URL="${DB_URL:-}"
DB_NAME="test"
LOGS_DIR_FROM_ENV="${LOGS_DIR:-}"
LOCAL_BACKUP_DIR="${LOCAL_BACKUP_DIR:-${HOME:-/root}/backups/cchoice}"
BACKUP_REMOTE="${BACKUP_REMOTE:-cchoice_backup:cchoice-assets/backups}"
RCLONE_ARGS="${RCLONE_ARGS:---s3-no-check-bucket}"

USE_LOCAL=false
DB_NAME_SET=false

while [[ $# -gt 0 ]]; do
    case "$1" in
        -a|--app-dir)
            if [[ -z "${2:-}" || "$2" =~ ^- ]]; then
                echo "Error: -a/--app-dir requires a directory" >&2
                exit 1
            fi
            APP_DIR="$2"
            shift 2
            ;;
        -d|--db)
            if [[ -z "${2:-}" || "$2" =~ ^- ]]; then
                echo "Error: -d/--db requires a value (prod or test)" >&2
                exit 1
            fi
            if [[ "$2" != "prod" && "$2" != "test" ]]; then
                echo "Error: -d/--db must be 'prod' or 'test'" >&2
                exit 1
            fi
            DB_NAME="$2"
            DB_NAME_SET=true
            DB_URL=""
            shift 2
            ;;
        -u|--db-url)
            if [[ -z "${2:-}" || "$2" =~ ^- ]]; then
                echo "Error: -u/--db-url requires a SQLite DB URL/path" >&2
                exit 1
            fi
            DB_URL="$2"
            DB_NAME_SET=false
            shift 2
            ;;
        -l|--local)
            USE_LOCAL=true
            if [[ -n "${2:-}" && ! "$2" =~ ^- ]]; then
                LOCAL_BACKUP_DIR="$2"
                shift
            fi
            shift
            ;;
        -r|--remote)
            if [[ -z "${2:-}" || "$2" =~ ^- ]]; then
                echo "Error: -r/--remote requires an rclone destination" >&2
                exit 1
            fi
            BACKUP_REMOTE="$2"
            shift 2
            ;;
        -h|--help)
            usage
            ;;
        *)
            echo "Unknown option: $1" >&2
            usage
            ;;
    esac
done

if ! APP_DIR="$(cd "$APP_DIR" && pwd)"; then
    echo "Error: app directory not found: $APP_DIR" >&2
    exit 1
fi
LOGS_DIR="${LOGS_DIR_FROM_ENV:-$APP_DIR/logs}"

if [[ -z "$DB_URL" && "$DB_NAME_SET" != true ]]; then
    load_db_url_from_env_file "$APP_DIR/.env"
fi
DB_URL="${DB_URL:-file:./${DB_NAME}.db}"

require_command sqlite3
require_command tar
if [[ "$USE_LOCAL" != true ]]; then
    require_command rclone
fi

TS="$(date +"%Y-%m-%d_%H-%M-%S")"
BACKUP_NAME="cchoice_backup_${TS}"
TMP_DIR="$(mktemp -d)"
ARCHIVE_ROOT="$TMP_DIR/$BACKUP_NAME"
ARCHIVE_PATH="$TMP_DIR/${BACKUP_NAME}.tar.gz"
SQLITE_BACKUP="$ARCHIVE_ROOT/cchoice.sqlite"

cleanup() {
    rm -rf "$TMP_DIR"
}
trap cleanup EXIT

mkdir -p "$ARCHIVE_ROOT"

DB_PATH="$(sqlite_db_path_from_url "$DB_URL" || true)"
if [[ -n "$DB_PATH" && ! -f "$DB_PATH" ]]; then
    echo "Error: database file not found: $DB_PATH" >&2
    exit 1
fi

echo "Backing up database: ${DB_PATH:-$DB_URL}"
(
    cd "$APP_DIR"
    sqlite3 "$DB_URL" ".backup '$SQLITE_BACKUP'"
)

if [[ -d "$LOGS_DIR" ]] && compgen -G "$LOGS_DIR/*.log" >/dev/null 2>&1; then
    echo "Found log files, adding to archive..."
    mkdir -p "$ARCHIVE_ROOT/logs"
    cp "$LOGS_DIR"/*.log "$ARCHIVE_ROOT/logs/"
else
    echo "No log files found, backing up database only..."
fi

echo "Creating tar archive..."
tar -czf "$ARCHIVE_PATH" -C "$TMP_DIR" "$BACKUP_NAME"

if [[ "$USE_LOCAL" = true ]]; then
    echo "Saving to local storage: $LOCAL_BACKUP_DIR"
    mkdir -p "$LOCAL_BACKUP_DIR"
    cp "$ARCHIVE_PATH" "$LOCAL_BACKUP_DIR/${BACKUP_NAME}.tar.gz"
else
    echo "Uploading to object storage: $BACKUP_REMOTE"
    # shellcheck disable=SC2086
    rclone copy "$ARCHIVE_PATH" "$BACKUP_REMOTE" $RCLONE_ARGS
fi

echo "Backup completed successfully: ${BACKUP_NAME}.tar.gz"
