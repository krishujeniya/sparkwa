#!/usr/bin/env bash
# SparkWA Lite — Binary Update Script
# Updates executable while preserving existing sessions, tokens, and config.
set -euo pipefail

if [ "$(id -u)" -ne 0 ]; then
    echo "Error: This script must be run as root (or via sudo)." >&2
    exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

NEW_BIN=""
if [ -f "$PROJECT_ROOT/bin/sparkwa" ]; then
    NEW_BIN="$PROJECT_ROOT/bin/sparkwa"
elif [ -f "$PROJECT_ROOT/sparkwa" ]; then
    NEW_BIN="$PROJECT_ROOT/sparkwa"
fi

if [ -z "$NEW_BIN" ]; then
    echo "Error: No new sparkwa binary found in $PROJECT_ROOT." >&2
    exit 1
fi

echo "Stopping sparkwa service..."
systemctl stop sparkwa.service

echo "Updating binary at /usr/local/bin/sparkwa..."
install -m 0755 "$NEW_BIN" /usr/local/bin/sparkwa

echo "Starting sparkwa service..."
systemctl start sparkwa.service

echo "✓ SparkWA Lite successfully updated to:"
/usr/local/bin/sparkwa version
systemctl status sparkwa.service --no-pager
