#!/usr/bin/env bash
# SparkWA — APT Repository Setup Script
# Enables 'apt install sparkwa' on Ubuntu and Debian systems.
# Author: Krish Ujeniya
set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}================================================================================${NC}"
echo -e "${BLUE}              SPARKWA APT REPOSITORY SETUP — By Krish Ujeniya                   ${NC}"
echo -e "${BLUE}================================================================================${NC}"

if [ "$(id -u)" -ne 0 ]; then
    echo -e "${RED}Error: This script must be run as root (or via sudo).${NC}" >&2
    exit 1
fi

if ! command -v apt-get >/dev/null 2>&1; then
    echo -e "${RED}Error: apt package manager not found. This script is for Debian/Ubuntu systems.${NC}" >&2
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

echo -e "Setting up SparkWA repository for ${GREEN}${TARGET_ARCH}${NC}..."

REPO_DIR="/var/lib/sparkwa/apt-repo"
mkdir -p "$REPO_DIR"

VERSION="1.0.0"
DEB_NAME="sparkwa_${VERSION}_${TARGET_ARCH}.deb"
DEB_URL="https://github.com/krishujeniya/sparkwa/releases/download/v${VERSION}/${DEB_NAME}"

# If release is already built locally in dist, use it; otherwise fetch from GitHub
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" 2>/dev/null && pwd || echo "")"
LOCAL_DIST="$SCRIPT_DIR/../dist/$DEB_NAME"

if [ -f "$LOCAL_DIST" ]; then
    echo -e "Using package: ${GREEN}${LOCAL_DIST}${NC}"
    cp "$LOCAL_DIST" "$REPO_DIR/$DEB_NAME"
else
    echo -e "Fetching package from: ${GREEN}${DEB_URL}${NC}"
    if command -v curl >/dev/null 2>&1; then
        curl -fsSL "$DEB_URL" -o "$REPO_DIR/$DEB_NAME" || true
    elif command -v wget >/dev/null 2>&1; then
        wget -q "$DEB_URL" -O "$REPO_DIR/$DEB_NAME" || true
    fi
fi

# Index repository if package exists
if [ -f "$REPO_DIR/$DEB_NAME" ]; then
    if ! command -v dpkg-scanpackages >/dev/null 2>&1; then
        apt-get update -qq >/dev/null 2>&1 || true
        apt-get install -y -qq dpkg-dev >/dev/null 2>&1 || true
    fi
    if command -v dpkg-scanpackages >/dev/null 2>&1; then
        (cd "$REPO_DIR" && dpkg-scanpackages . /dev/null 2>/dev/null | gzip -9c > Packages.gz)
    fi
fi

# Configure APT source list
SOURCES_LIST="/etc/apt/sources.list.d/sparkwa.list"
echo "deb [trusted=yes] file:$REPO_DIR ./" > "$SOURCES_LIST"

echo -e "Updating package index..."
apt-get update -o Dir::Etc::sourcelist="$SOURCES_LIST" -o Dir::Etc::sourceparts="-" -o APT::Get::List-Cleanup="0" -qq 2>/dev/null || apt-get update -qq

echo -e "${GREEN}✓ SparkWA APT repository configured successfully.${NC}"
echo -e ""
echo -e "You can now install SparkWA anytime by running:"
echo -e "  ${GREEN}sudo apt install -y sparkwa${NC}"
echo -e "${BLUE}================================================================================${NC}"
