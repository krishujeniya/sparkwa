# SparkWA Lite

[![Go Version](https://img.shields.io/badge/Go-1.23%2B-blue.svg)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Ubuntu%20%2F%20Debian%20Linux-orange.svg)]()
[![Memory Footprint](https://img.shields.io/badge/RAM-~30%20MB-brightgreen.svg)]()

**SparkWA Lite** is an ultra-lightweight, production-grade, headless bridge that connects a dedicated **WhatsApp** account to your existing **Google Gemini Spark** agent through Google's documented Gmail-triggered task schedules.

Designed specifically for headless Ubuntu VPS deployments, it runs as a single native Linux binary under `systemd` with zero browser, desktop, Node.js, Python, or container dependencies.

---

## Architecture

```
WhatsApp User (Number 2)
       │ (Encrypted WhatsApp Text)
       ▼
WhatsApp Bot Account (Number 1)
       │ (whatsmeow Multi-Device)
       ▼
┌─────────────────────────────────────────────────────────────┐
│                      SPARKWA LITE                           │
│  • Fail-Closed Sender Verifier (Only Number 2 Allowed)      │
│  • SQLite Task Queue & Recovery Engine (PRAGMA WAL)        │
│  • Structured [SPARKWA:REQ_ID] Protocol Formatter          │
│  • Least-Privilege Google Gmail API OAuth 2.0 Engine        │
└──────────────────────────────┬──────────────────────────────┘
                               │ (Outbound HTTPS / OAuth 2.0)
                               ▼
┌─────────────────────────────────────────────────────────────┐
│                    GOOGLE CLUSTER                           │
│  Bridge Gmail Inbox ────[Email Command]───► Spark Inbox     │
│                                                │            │
│                                                ▼            │
│                                       Gemini Spark Agent    │
│                                       (Task & Skills)       │
│                                                │            │
│  Bridge Gmail Inbox ◄───[Reply/Draft]──────────┘            │
└──────────────────────────────┬──────────────────────────────┘
                               │ (45s Configurable Poller)
                               ▼
WhatsApp User (Number 2) ◄─────┴── Verified Result Delivered
```

---

## Key Features

- **Single Native Binary:** Built in Go. Zero external runtime dependencies.
- **Ultra-Low Resource Footprint:** Uses ~30 MB RAM and 0.0% idle CPU. Runs comfortably on 512 MB to 1 GB VPS instances.
- **Headless SSH QR Pairing:** Displays an ASCII QR code directly inside your SSH terminal to link WhatsApp Number 1 via Linked Devices.
- **Fail-Closed Security:** Strict allowlist for Phone Number 2. Unregistered numbers, groups, broadcasts, and self-messages are silently dropped.
- **Official Google OAuth 2.0:** Uses least-privilege Gmail scopes (`gmail.send`, `gmail.readonly`) with automatic background token refresh.
- **Headless OAuth Helper:** Includes built-in loopback authorization that works seamlessly over SSH port forwarding.
- **Preserves Your Existing Spark Agent:** Integrates directly with your Google-hosted Gemini Spark agent; does not replace it with an ordinary model API chatbot.
- **Persistent Task Queue:** Embedded SQLite database with automatic crash recovery, duplicate message suppression, and timeout management.
- **Approval Guardrail Detection:** Automatically detects when Spark pauses for confirmation and forwards the approval prompt to WhatsApp.
- **Hardened systemd Unit:** Runs as a dedicated non-root `sparkwa` service account with strict Linux filesystem and kernel isolation.

---

## Capability Feasibility Gates

SparkWA Lite enforces explicit capability states to maintain transparency regarding what is verified versus unverified:

| Capability State | Meaning |
|------------------|---------|
| `NOT_CONFIGURED` | Credentials or pairing missing. |
| `GMAIL_CONNECTED` | Gmail API authenticated; Spark monitor not yet confirmed. |
| `SPARK_TRIGGER_CONFIGURED` | Task dispatch confirmed; awaiting return path verification. |
| `RESPONSE_PATH_VERIFIED` | Return email or draft discovered and parsed successfully. |
| `SUBMISSION_ONLY` | One-way bridge confirmed; Spark executes actions in Google Workspace without chat replies. |
| `FULLY_VERIFIED` | Complete round-trip verified end-to-end with live Spark tasks. |

---

## Quick Start (VPS Deployment)

### 1. Upload Release Archive to VPS
```bash
scp sparkwa-lite-v1.0.0-linux-amd64.tar.gz credentials.json user@your-vps-ip:/tmp/
```

### 2. Extract and Run Automated Installer
```bash
ssh user@your-vps-ip
cd /tmp && tar -xzf sparkwa-lite-v1.0.0-linux-amd64.tar.gz
cd sparkwa-lite && sudo ./scripts/install.sh
```

### 3. Move Credentials & Edit Config
```bash
sudo mv /tmp/credentials.json /etc/sparkwa-lite/credentials.json
sudo chmod 0600 /etc/sparkwa-lite/credentials.json
sudo nano /etc/sparkwa-lite/config.env
```

### 4. Authenticate Gmail and Pair WhatsApp
```bash
# Authorize Gmail API (via loopback or SSH tunnel port 8085)
sudo -u sparkwa sparkwa gmail-auth --config /etc/sparkwa-lite/config.env

# Pair WhatsApp Number 1 via Terminal QR Code
sudo -u sparkwa sparkwa login --config /etc/sparkwa-lite/config.env
```

### 5. Validate Health and Start Service
```bash
# Run system doctor
sudo -u sparkwa sparkwa doctor --config /etc/sparkwa-lite/config.env

# Start systemd daemon
sudo systemctl start sparkwa
sudo journalctl -u sparkwa -f
```

---

## WhatsApp Commands

Once running, send these commands from **Number 2** to **Number 1**:

- `/help` — Display command guide.
- `/status` — View bridge, WhatsApp, and Gmail health.
- `/jobs` — List the 5 most recent tasks.
- `/job <id>` — Show full details of a specific task.
- `/cancel <id>` — Cancel a queued or active task.
- `/ping` — Verify bridge responsiveness.
- **Any other text message** — Automatically dispatched to Gemini Spark as an agent task.

---

## Documentation

Detailed documentation is available in the `docs/` folder:

- [docs/SPARK_SETUP.md](docs/SPARK_SETUP.md) — Exact setup for Gemini Spark Gmail monitors, filters, and prompt copy-paste.
- [docs/GMAIL_OAUTH_SETUP.md](docs/GMAIL_OAUTH_SETUP.md) — Google Cloud Console OAuth 2.0 credentials and headless setup.
- [docs/VPS_INSTALL.md](docs/VPS_INSTALL.md) — Step-by-step Ubuntu VPS deployment guide.
- [docs/TROUBLESHOOTING.md](docs/TROUBLESHOOTING.md) — Common error symptoms, diagnostics, and resolutions.
- [docs/SECURITY.md](docs/SECURITY.md) — Threat model, security boundaries, and WhatsApp risk disclosures.
- [docs/VERIFICATION.md](docs/VERIFICATION.md) — Test suite results, acceptance gates, and benchmark measurements.

---

## License

This project is licensed under the [MIT License](LICENSE).
