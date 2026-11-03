#!/usr/bin/env bash
# SparkWA — State and Configuration Backup Script
set -euo pipefail

if [ "$(id -u)" -ne 0 ]; then
    echo "Error: This script must be run as root (or via sudo)." >&2
    exit 1
fi

TIMESTAMP=$(date -u +"%Y%m%d_%H%M%S")
BACKUP_DIR="/var/backups/sparkwa"
BACKUP_FILE="$BACKUP_DIR/sparkwa-backup-$TIMESTAMP.tar.gz"

mkdir -p "$BACKUP_DIR"
chmod 0700 "$BACKUP_DIR"

echo "Creating atomic SQLite database snapshot..."
# Safely snapshot sqlite database if running
if command -v sqlite3 >/dev/null 2>&1; then
    if [ -f /var/lib/sparkwa/sparkwa.db ]; then
        sqlite3 /var/lib/sparkwa/sparkwa.db ".backup /var/lib/sparkwa/sparkwa.db.bak"
    elif [ -f /var/lib/sparkwa-lite/sparkwa.db ]; then
        sqlite3 /var/lib/sparkwa-lite/sparkwa.db ".backup /var/lib/sparkwa-lite/sparkwa.db.bak"
    fi
fi

ARCHIVE_PATHS=()
[ -d /etc/sparkwa ] && ARCHIVE_PATHS+=("etc/sparkwa")
[ -d /var/lib/sparkwa ] && ARCHIVE_PATHS+=("var/lib/sparkwa")
[ -d /etc/sparkwa-lite ] && ARCHIVE_PATHS+=("etc/sparkwa-lite")
[ -d /var/lib/sparkwa-lite ] && ARCHIVE_PATHS+=("var/lib/sparkwa-lite")

if [ ${#ARCHIVE_PATHS[@]} -eq 0 ]; then
    echo "Warning: No sparkwa configuration or data directories found in /etc or /var/lib."
else
    echo "Archiving ${ARCHIVE_PATHS[*]}..."
    tar -czf "$BACKUP_FILE" \
        --exclude="*.bak" \
        -C / \
        "${ARCHIVE_PATHS[@]}"
fi

rm -f /var/lib/sparkwa/sparkwa.db.bak /var/lib/sparkwa-lite/sparkwa.db.bak 2>/dev/null || true
[ -f "$BACKUP_FILE" ] && chmod 0600 "$BACKUP_FILE"

echo "✓ Backup created successfully at:"
echo "  $BACKUP_FILE"
ls -lh "$BACKUP_FILE"
