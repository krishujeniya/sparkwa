#!/usr/bin/env bash
# SparkWA — Uninstaller
set -euo pipefail

if [ "$(id -u)" -ne 0 ]; then
    echo "Error: This script must be run as root (or via sudo)." >&2
    exit 1
fi

echo "Stopping and disabling sparkwa.service..."
systemctl stop sparkwa.service 2>/dev/null || true
systemctl disable sparkwa.service 2>/dev/null || true

echo "Removing systemd service unit..."
rm -f /etc/systemd/system/sparkwa.service
systemctl daemon-reload

echo "Removing binary..."
rm -f /usr/local/bin/sparkwa

echo ""
read -p "Do you want to delete all configuration and session data (/var/lib/sparkwa and /etc/sparkwa)? [y/N]: " CONFIRM
if [[ "$CONFIRM" =~ ^[Yy]$ ]]; then
    echo "Removing /var/lib/sparkwa and /etc/sparkwa..."
    rm -rf /var/lib/sparkwa /etc/sparkwa
    userdel sparkwa 2>/dev/null || true
    echo "All data removed."
else
    echo "Preserved configuration and session data in /var/lib/sparkwa and /etc/sparkwa."
fi

echo "✓ SparkWA has been cleanly uninstalled."
