#!/bin/bash
set -e

###
# CONFIGURATION (edit these for your environment)
###

CCHOICE_DIR="/root/cchoice"
DB_PATH="$CCHOICE_DIR/test.db"
LOGS_DIR="$CCHOICE_DIR/logs"
TS=$(date +"%Y-%m-%d_%H-%M-%S")
TMP="/tmp/db_backup_${TS}.tar.gz"
BUCKET_REMOTE="cchoice_backup:cchoice-assets/backups"

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
# UPLOAD TO OBJECT STORAGE
# This stores each backup as a new file: backup_YYYY-mm-dd_HH-MM-SS.tar.gz
###
echo "Uploading to object storage..."
rclone copy "$TMP" "$BUCKET_REMOTE" --s3-no-check-bucket

###
# CLEANUP TEMP FILES
###
echo "Cleaning up temporary files..."
rm -f "$TMP" "$SQLITE_BACKUP"

echo "Backup completed successfully: backup_${TS}.tar.gz"
