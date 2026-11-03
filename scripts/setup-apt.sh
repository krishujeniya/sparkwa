#!/usr/bin/env bash
# SparkWA — Automated Installer & Runner
set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
BOLD='\033[1m'
NC='\033[0m'

echo -e "${BOLD}================================================================================${NC}"
echo -e "${BOLD}                     SPARKWA AUTOMATED INSTALLER & RUNNER                       ${NC}"
echo -e "${BOLD}================================================================================${NC}"

if [ "$(id -u)" -ne 0 ]; then
    echo -e "${RED}Error: This script must be run as root (or via sudo).${NC}" >&2
    exit 1
fi

ARCH=$(uname -m)
case "$ARCH" in
    x86_64)
        TARGET_ARCH="amd64"
        ;;
    aarch64|arm64)
        TARGET_ARCH="arm64"
        ;;
    *)
        echo -e "${RED}Error: Unsupported architecture: $ARCH. SparkWA supports amd64 and arm64.${NC}" >&2
        exit 1
        ;;
esac

echo -e "Installing SparkWA for ${GREEN}${TARGET_ARCH}${NC}..."

VERSION="1.0.0"
DEB_NAME="sparkwa_${VERSION}_${TARGET_ARCH}.deb"
DEB_URL="https://github.com/krishujeniya/sparkwa/releases/download/v${VERSION}/${DEB_NAME}"
TAR_URL="https://github.com/krishujeniya/sparkwa/releases/download/v${VERSION}/sparkwa-v${VERSION}-linux-${TARGET_ARCH}.tar.gz"

TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

INSTALLED=0

# Method A: Install official Debian package via apt
if command -v apt-get >/dev/null 2>&1; then
    SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" 2>/dev/null && pwd || echo "")"
    LOCAL_DEB="$SCRIPT_DIR/../dist/$DEB_NAME"

    if [ -f "$LOCAL_DEB" ]; then
        cp "$LOCAL_DEB" "$TMP_DIR/$DEB_NAME"
    elif command -v curl >/dev/null 2>&1; then
        curl -fsSL "$DEB_URL" -o "$TMP_DIR/$DEB_NAME" 2>/dev/null || true
    elif command -v wget >/dev/null 2>&1; then
        wget -q "$DEB_URL" -O "$TMP_DIR/$DEB_NAME" 2>/dev/null || true
    fi

    if [ -f "$TMP_DIR/$DEB_NAME" ]; then
        echo -e "Installing package via apt..."
        apt-get update -qq >/dev/null 2>&1 || true
        apt-get install -y -qq "$TMP_DIR/$DEB_NAME" >/dev/null 2>&1 || dpkg -i "$TMP_DIR/$DEB_NAME"
        INSTALLED=1
    fi
fi

# Method B: Standalone Binary Fallback
if [ "$INSTALLED" -ne 1 ]; then
    echo -e "Downloading binary release..."
    if command -v curl >/dev/null 2>&1; then
        curl -fsSL "$TAR_URL" -o "$TMP_DIR/sparkwa.tar.gz"
    elif command -v wget >/dev/null 2>&1; then
        wget -q "$TAR_URL" -O "$TMP_DIR/sparkwa.tar.gz"
    fi

    tar -xzf "$TMP_DIR/sparkwa.tar.gz" -C "$TMP_DIR"
    install -m 0755 "$TMP_DIR/sparkwa/sparkwa" /usr/local/bin/sparkwa

    # Setup user & paths
    if ! id "sparkwa" >/dev/null 2>&1; then
        useradd --system --no-create-home --user-group --shell /usr/sbin/nologin sparkwa || true
    fi
    mkdir -p /etc/sparkwa /var/lib/sparkwa
    chown -R sparkwa:sparkwa /var/lib/sparkwa
    chmod 0700 /var/lib/sparkwa

    if [ -f "$TMP_DIR/sparkwa/deploy/sparkwa.service" ]; then
        cp "$TMP_DIR/sparkwa/deploy/sparkwa.service" /etc/systemd/system/
        systemctl daemon-reload || true
    fi
    INSTALLED=1
fi

echo -e "${GREEN}✓ SparkWA successfully installed.${NC}"
echo -e ""
echo -e "Launching interactive control dashboard..."
sleep 1

# Launch directly attached to controlling terminal
if [ -e /dev/tty ]; then
    exec /usr/local/bin/sparkwa </dev/tty >/dev/tty 2>&1
else
    exec /usr/local/bin/sparkwa
fi
