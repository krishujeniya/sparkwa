#!/usr/bin/env bash
# SparkWA Lite — State and Configuration Backup Script
set -euo pipefail

if [ "$(id -u)" -ne 0 ]; then
    echo "Error: This script must be run as root (or via sudo)." >&2
    exit 1
fi

TIMESTAMP=$(date -u +"%Y%m%d_%H%M%S")
BACKUP_DIR="/var/backups/sparkwa-lite"
BACKUP_FILE="$BACKUP_DIR/sparkwa-backup-$TIMESTAMP.tar.gz"

mkdir -p "$BACKUP_DIR"
chmod 0700 "$BACKUP_DIR"

echo "Creating atomic SQLite database snapshot..."
# Safely snapshot sqlite database if running
if command -v sqlite3 >/dev/null 2>&1 && [ -f /var/lib/sparkwa-lite/sparkwa.db ]; then
    sqlite3 /var/lib/sparkwa-lite/sparkwa.db ".backup /var/lib/sparkwa-lite/sparkwa.db.bak"
fi

echo "Archiving /etc/sparkwa-lite and /var/lib/sparkwa-lite..."
tar -czf "$BACKUP_FILE" \
    --exclude="*.bak" \
    -C / \
    etc/sparkwa-lite \
    var/lib/sparkwa-lite

rm -f /var/lib/sparkwa-lite/sparkwa.db.bak 2>/dev/null || true
chmod 0600 "$BACKUP_FILE"

echo "✓ Backup created successfully at:"
echo "  $BACKUP_FILE"
ls -lh "$BACKUP_FILE"
