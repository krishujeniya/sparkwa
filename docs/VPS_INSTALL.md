# Ubuntu VPS Installation & Deployment Guide

This guide describes the complete workflow to deploy **SparkWA** to a headless Ubuntu VPS (Ubuntu 24.04 LTS / 22.04 LTS / Debian).

The application runs headlessly as a native background service using systemd.

---

## 1. System Requirements

- **Operating System:** Ubuntu 24.04 LTS or 22.04 LTS (x86_64 or aarch64).
- **Compute:** 1 vCPU, 512 MB to 1 GB RAM.
- **Disk:** 50 MB free disk space.
- **Network:** Direct outbound HTTPS (port 443) and WhatsApp TCP (ports 443 / 5222).
- **Runtime:** Zero dependencies (No Node.js, Python, Docker, Chromium, or desktop required).

---

## 2. Step 1: Upload Archive to VPS

From your local machine where you built or downloaded the release archive:

```bash
# Upload release archive and credentials JSON to your VPS
scp sparkwa-v1.0.0-linux-amd64.tar.gz user@your-vps-ip:/tmp/
scp credentials.json user@your-vps-ip:/tmp/
```

---

## 3. Step 2: SSH into VPS and Extract

```bash
ssh user@your-vps-ip

# Navigate to /tmp and extract package
cd /tmp
tar -xzf sparkwa-v1.0.0-linux-amd64.tar.gz
cd sparkwa
```

---

## 4. Step 3: Run the Automated Installer

Run the installer with root privileges:

```bash
sudo ./scripts/install.sh
```

The installer automatically:
1. Detects system architecture (`amd64` or `arm64`).
2. Creates the dedicated non-login system user `sparkwa`.
3. Installs the single static binary to `/usr/local/bin/sparkwa`.
4. Creates protected directory `/var/lib/sparkwa` with mode `0700`.
5. Creates configuration directory `/etc/sparkwa` with mode `0750`.
6. Installs and enables the hardened `sparkwa.service` systemd unit.

---

## 5. Step 4: Configure Phone Numbers & Gmail

Move your Google Cloud OAuth credentials into place:

```bash
sudo mv /tmp/credentials.json /etc/sparkwa/credentials.json
sudo chown root:sparkwa /etc/sparkwa/credentials.json
sudo chmod 0600 /etc/sparkwa/credentials.json
```

Edit the configuration file:

```bash
sudo nano /etc/sparkwa/config.env
```

Ensure the following variables are set:

```env
# E.164 phone numbers
BOT_PHONE_NUMBER="+1234567890"        # Number 1 (the bot account)
AUTHORIZED_PHONE_NUMBER="+1987654321" # Number 2 (the only allowed user)

# Gmail Mailboxes
BRIDGE_GMAIL_ADDRESS="bot-bridge@gmail.com"
SPARK_GMAIL_ADDRESS="personal-spark@gmail.com"

# Paths
DATA_DIR="/var/lib/sparkwa"
GMAIL_CREDENTIALS_FILE="/etc/sparkwa/credentials.json"
GMAIL_TOKEN_FILE="/var/lib/sparkwa/token.json"

# Timing
GMAIL_POLL_INTERVAL="45s"
TASK_TIMEOUT="15m"
LOG_LEVEL="info"
```

Save and exit (`Ctrl+O`, `Enter`, `Ctrl+X`).

---

## 6. Step 5: Authorize Gmail API (OAuth 2.0)

Run the Gmail authorization helper as the `sparkwa` service user:

```bash
sudo -u sparkwa sparkwa gmail-auth --config /etc/sparkwa/config.env
```

Follow the on-screen instructions (see [GMAIL_OAUTH_SETUP.md](GMAIL_OAUTH_SETUP.md) for details):
- Either forward port 8085 (`ssh -L 8085:127.0.0.1:8085 user@your-vps-ip`)
- Or paste the redirect code directly into the terminal prompt.

---

## 7. Step 6: Pair WhatsApp Number 1 (Terminal QR Code)

Pair the WhatsApp account for Number 1 directly in your SSH terminal:

```bash
sudo -u sparkwa sparkwa login --config /etc/sparkwa/config.env
```

1. An ASCII QR code will be displayed directly inside your SSH terminal.
2. Open WhatsApp on phone **Number 1**.
3. Tap **Settings** > **Linked Devices** > **Link a Device**.
4. Point your camera at the SSH terminal QR code.
5. The terminal will output:
   ```text
   ✓ WhatsApp device pairing successful!
   ✓ WhatsApp pairing complete. Session saved in SQLite database.
   ```

---

## 8. Step 7: Run System Diagnostics (Doctor)

Run the diagnostic check to ensure all subsystems are ready:

```bash
sudo -u sparkwa sparkwa doctor --config /etc/sparkwa/config.env
```

Verify that all critical checks report `[✓ PASS]`:
- Configuration Validation
- Filesystem Permissions (0700 data dir, 0600 token & config)
- SQLite Database Integrity
- WhatsApp Pairing Status
- Gmail API Connectivity

---

## 9. Step 8: Start the systemd Background Service

Start the service and enable it to persist across VPS reboots:

```bash
sudo systemctl start sparkwa
sudo systemctl status sparkwa
```

Check live logs via `journalctl`:

```bash
sudo journalctl -u sparkwa -f
```

---

## 10. Step 9: Send Your First WhatsApp Command

From phone **Number 2**, send a WhatsApp message to phone **Number 1**:

```text
/ping
```

The bot will reply within seconds:
```text
🏓 *Pong!*

SparkWA is responsive.
Timestamp: 2026-10-31T18:00:00Z UTC
```

Next, test status:
```text
/status
```

Then submit your first task:
```text
Please summarize my upcoming meetings for this week.
```

SparkWA will return:
```text
*SparkWA* [SUBMITTED]

*ID:* `d3b07384`
*Task:* Please summarize my upcoming meetings for this week.

_Forwarded to Gemini Spark via Gmail monitor. Waiting for agent processing._
```

Once your Gemini Spark agent processes the command, the result is forwarded back to your WhatsApp chat!

---

## 11. Maintenance and Useful Commands

```bash
# View live logs
sudo journalctl -u sparkwa -f

# Check recent tasks and capability status
sudo -u sparkwa sparkwa status --config /etc/sparkwa/config.env

# Stop service
sudo systemctl stop sparkwa

# Restart service
sudo systemctl restart sparkwa

# Create automated backup
sudo /usr/local/bin/sparkwa-backup # (or /tmp/sparkwa/scripts/backup.sh)
```
