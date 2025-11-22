#!/bin/bash
set -e

###
# CONFIGURATION (edit these for your environment)
###

DB_PATH="./test.db"
LOG_PATH="./logs/*.log"
TS=$(date +"%Y-%m-%d_%H-%M-%S")
TMP="/tmp/db_backup_${TS}.tar.gz"

BUCKET_REMOTE="cchoice_backup:cchoice-assets/backups"


###
# SQLite SAFE BACKUP
# This avoids copying the DB while it's being written
###
SQLITE_BACKUP="/tmp/sqlite_backup_${TS}.sqlite"
sqlite3 "$DB_PATH" ".backup '$SQLITE_BACKUP'"


###
# PACKAGE FILES INTO TAR.GZ
# (DB + logs)
###
tar -czf "$TMP" "$SQLITE_BACKUP" $LOG_PATH


###
# UPLOAD TO OBJECT STORAGE
# This stores each backup as a new file: backup_YYYY-mm-dd_HH-MM-SS.tar.gz
###
rclone copy "$TMP" "$BUCKET_REMOTE/backup_${TS}.tar.gz" --s3-no-check-bucket


###
# CLEANUP TEMP FILES
###
rm -f "$TMP" "$SQLITE_BACKUP"
