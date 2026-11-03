#!/usr/bin/env bash
# SparkWA — Production VPS Installer
# Target: Ubuntu 24.04 LTS / Debian / Linux
set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}================================================================================${NC}"
echo -e "${BLUE}                       SPARKWA VPS INSTALLER                                    ${NC}"
echo -e "${BLUE}================================================================================${NC}"

# 1. Root check
if [ "$(id -u)" -ne 0 ]; then
    echo -e "${RED}Error: This script must be run as root (or via sudo).${NC}" >&2
    exit 1
fi

# 2. OS Compatibility Check
if [ -f /etc/os-release ]; then
    . /etc/os-release
    echo -e "Detected Operating System: ${GREEN}${PRETTY_NAME:-Linux}${NC}"
else
    echo -e "${YELLOW}Warning: /etc/os-release not found. Assuming standard Linux.${NC}"
fi

# 3. Architecture Detection
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
echo -e "Detected Architecture: ${GREEN}${TARGET_ARCH}${NC} ($ARCH)"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# 4. Locate Binary
BIN_SOURCE=""
if [ -f "$PROJECT_ROOT/bin/sparkwa" ]; then
    BIN_SOURCE="$PROJECT_ROOT/bin/sparkwa"
elif [ -f "$PROJECT_ROOT/sparkwa" ]; then
    BIN_SOURCE="$PROJECT_ROOT/sparkwa"
elif command -v go >/dev/null 2>&1; then
    echo -e "Compiling sparkwa binary for linux/$TARGET_ARCH..."
    (cd "$PROJECT_ROOT" && CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/sparkwa ./cmd/sparkwa)
    BIN_SOURCE="$PROJECT_ROOT/bin/sparkwa"
else
    echo -e "${RED}Error: Could not find prebuilt sparkwa binary and Go compiler is not installed.${NC}" >&2
    exit 1
fi

# 5. Create Dedicated System User
USER_NAME="sparkwa"
GROUP_NAME="sparkwa"

if ! id "$USER_NAME" >/dev/null 2>&1; then
    echo -e "Creating system service user: ${GREEN}${USER_NAME}${NC}..."
    useradd --system --no-create-home --user-group --shell /usr/sbin/nologin "$USER_NAME"
else
    echo -e "Service user ${GREEN}${USER_NAME}${NC} already exists."
fi

# 6. Create Protected Directories
CONFIG_DIR="/etc/sparkwa"
DATA_DIR="/var/lib/sparkwa"
BIN_DEST="/usr/local/bin/sparkwa"

# Auto-migrate legacy sparkwa-lite paths if present
if [ ! -d "$CONFIG_DIR" ] && [ -d "/etc/sparkwa-lite" ]; then
    echo -e "Migrating legacy config directory /etc/sparkwa-lite to ${GREEN}${CONFIG_DIR}${NC}..."
    cp -rn /etc/sparkwa-lite "$CONFIG_DIR"
fi
if [ ! -d "$DATA_DIR" ] && [ -d "/var/lib/sparkwa-lite" ]; then
    echo -e "Migrating legacy data directory /var/lib/sparkwa-lite to ${GREEN}${DATA_DIR}${NC}..."
    cp -rn /var/lib/sparkwa-lite "$DATA_DIR"
fi

echo -e "Setting up directory structure..."
mkdir -p "$CONFIG_DIR"
mkdir -p "$DATA_DIR"

# 7. Install Binary
echo -e "Installing binary to ${GREEN}${BIN_DEST}${NC}..."
install -m 0755 "$BIN_SOURCE" "$BIN_DEST"

# 8. Setup Configuration Template (Preserve existing!)
CONFIG_FILE="$CONFIG_DIR/config.env"
if [ ! -f "$CONFIG_FILE" ]; then
    echo -e "Installing default configuration template to ${GREEN}${CONFIG_FILE}${NC}..."
    if [ -f "$PROJECT_ROOT/.env.example" ]; then
        cp "$PROJECT_ROOT/.env.example" "$CONFIG_FILE"
    else
        cat << 'EOF' > "$CONFIG_FILE"
# ==============================================================================
# SPARKWA CONFIGURATION
# ==============================================================================

# WhatsApp Phone Numbers (Must be in standard E.164 format: +<country><digits>)
BOT_PHONE_NUMBER="+1234567890"         # Phone Number 1 (the bot account)
AUTHORIZED_PHONE_NUMBER="+1987654321"  # Phone Number 2 (the ONLY permitted user)

# Google Gmail Accounts
BRIDGE_GMAIL_ADDRESS="bot-bridge@gmail.com"         # Account used by SparkWA bridge
SPARK_GMAIL_ADDRESS="personal-spark@gmail.com"      # Google account where Gemini Spark runs

# Paths & Security
DATA_DIR="/var/lib/sparkwa"
GMAIL_CREDENTIALS_FILE="/etc/sparkwa/credentials.json"
GMAIL_TOKEN_FILE="/var/lib/sparkwa/token.json"

# Polling & Quota Management
GMAIL_POLL_INTERVAL="45s"
TASK_TIMEOUT="15m"
MAX_CONCURRENT_JOBS="5"
LOG_LEVEL="info"
EOF
    fi
    chmod 0600 "$CONFIG_FILE"
    chown "$USER_NAME:$GROUP_NAME" "$CONFIG_FILE"
else
    echo -e "Preserving existing configuration file: ${GREEN}${CONFIG_FILE}${NC}"
fi

# Set directory permissions
chown -R "$USER_NAME:$GROUP_NAME" "$DATA_DIR"
chmod 0700 "$DATA_DIR"
chown root:"$GROUP_NAME" "$CONFIG_DIR"
chmod 0750 "$CONFIG_DIR"

# 9. Install Systemd Service
SERVICE_SRC="$PROJECT_ROOT/deploy/sparkwa.service"
SERVICE_DEST="/etc/systemd/system/sparkwa.service"

if [ -f "$SERVICE_SRC" ]; then
    echo -e "Installing systemd service unit to ${GREEN}${SERVICE_DEST}${NC}..."
    cp "$SERVICE_SRC" "$SERVICE_DEST"
    chmod 0644 "$SERVICE_DEST"
    systemctl daemon-reload
    systemctl enable sparkwa.service
fi

echo -e "\n${GREEN}================================================================================${NC}"
echo -e "${GREEN}                     SPARKWA INSTALLATION COMPLETE                              ${NC}"
echo -e "${GREEN}================================================================================${NC}"
echo -e "NEXT STEPS TO GO LIVE:"
echo -e "1. Edit your configuration file:"
echo -e "   ${YELLOW}sudo nano /etc/sparkwa/config.env${NC}"
echo -e "2. Upload Google Cloud OAuth credentials JSON to:"
echo -e "   ${YELLOW}/etc/sparkwa/credentials.json${NC} (mode 0600)"
echo -e "3. Authenticate with Google Gmail API:"
echo -e "   ${YELLOW}sudo -u sparkwa sparkwa gmail-auth --config /etc/sparkwa/config.env${NC}"
echo -e "4. Pair WhatsApp Number 1 via Terminal QR code:"
echo -e "   ${YELLOW}sudo -u sparkwa sparkwa login --config /etc/sparkwa/config.env${NC}"
echo -e "5. Validate all system health checks:"
echo -e "   ${YELLOW}sudo -u sparkwa sparkwa doctor --config /etc/sparkwa/config.env${NC}"
echo -e "6. Start the headless background service:"
echo -e "   ${YELLOW}sudo systemctl start sparkwa${NC}"
echo -e "7. Check live service logs:"
echo -e "   ${YELLOW}sudo journalctl -u sparkwa -f${NC}"
echo -e "${GREEN}================================================================================${NC}"
