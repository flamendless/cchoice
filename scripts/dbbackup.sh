#!/bin/bash
set -e

usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -l, --local [DIR]    Save backup to local storage instead of object storage"
    echo "                       Default local directory: /root/cchoice/backups"
    echo "  -h, --help           Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0                   # Backup to Linode object storage"
    echo "  $0 --local           # Backup to default local directory"
    echo "  $0 --local /path/to/backups  # Backup to custom local directory"
    exit 0
}

###
# CONFIGURATION (edit these for your environment)
###

CCHOICE_DIR="/root/cchoice"
DB_PATH="$CCHOICE_DIR/test.db"
LOGS_DIR="$CCHOICE_DIR/logs"
TS=$(date +"%Y-%m-%d_%H-%M-%S")
TMP="/tmp/db_backup_${TS}.tar.gz"
BUCKET_REMOTE="cchoice_backup:cchoice-assets/backups"
LOCAL_BACKUP_DIR="/root/backups"

USE_LOCAL=false
while [[ $# -gt 0 ]]; do
    case $1 in
        -l|--local)
            USE_LOCAL=true
            if [[ -n "$2" && ! "$2" =~ ^- ]]; then
                LOCAL_BACKUP_DIR="$2"
                shift
            fi
            shift
            ;;
        -h|--help)
            usage
            ;;
        *)
            echo "Unknown option: $1"
            usage
            ;;
    esac
done

###
# SQLite SAFE BACKUP
# This avoids copying the DB while it's being written
###
SQLITE_BACKUP="/tmp/sqlite_backup_${TS}.sqlite"
echo "Creating SQLite backup..."
sqlite3 "$DB_PATH" ".backup '$SQLITE_BACKUP'"

###
# PACKAGE FILES INTO TAR.GZ
# (DB + logs, if they exist)
###
echo "Creating tar archive..."
FILES_TO_BACKUP=("$SQLITE_BACKUP")
if [ -d "$LOGS_DIR" ] && compgen -G "$LOGS_DIR/*.log" > /dev/null 2>&1; then
    echo "Found log files, adding to archive..."
    FILES_TO_BACKUP+=("$LOGS_DIR"/*.log)
else
    echo "No log files found, backing up database only..."
fi

tar -czf "$TMP" "${FILES_TO_BACKUP[@]}"

###
# SAVE BACKUP (local or object storage)
###
if [ "$USE_LOCAL" = true ]; then
    echo "Saving to local storage: $LOCAL_BACKUP_DIR"
    mkdir -p "$LOCAL_BACKUP_DIR"
    cp "$TMP" "$LOCAL_BACKUP_DIR/backup_${TS}.tar.gz"
else
    echo "Uploading to object storage..."
    rclone copy "$TMP" "$BUCKET_REMOTE" --s3-no-check-bucket
fi

###
# CLEANUP TEMP FILES
###
echo "Cleaning up temporary files..."
rm -f "$TMP" "$SQLITE_BACKUP"

echo "Backup completed successfully: backup_${TS}.tar.gz"
