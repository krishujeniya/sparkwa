```text
╔════════════════════════════════════════════════════════════════════════════════╗
║  ███████╗██████╗  █████╗ ██████╗ ██╗  ██╗   ██╗    ██╗ █████╗                  ║
║  ██╔════╝██╔══██╗██╔══██╗██╔══██╗██║ ██╔╝   ██║    ██║██╔══██╗                 ║
║  ███████╗██████╔╝███████║██████╔╝█████╔╝    ██║ █╗ ██║███████║                  ║
║  ╚════██║██╔═══╝ ██╔══██║██╔══██╗██╔═██╗    ██║███╗██║██╔══██║                 ║
║  ███████║██║     ██║  ██║██║  ██║██║  ██╗   ╚███╔███╔╝██║  ██║                 ║
║  ╚══════╝╚═╝     ╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═╝    ╚══╝╚══╝ ╚═╝  ╚═╝                 ║
║                                                                                ║
║  THE HEADLESS WHATSAPP TO GEMINI SPARK BRIDGE                                  ║
║  Pure Go • Single Native Binary • Zero Dependencies • Production Hardened      ║
╚════════════════════════════════════════════════════════════════════════════════╝
```

# SparkWA

[![Go Version](https://img.shields.io/badge/Go-1.23%2B-black.svg?style=for-the-badge&logo=go)](https://golang.org)
[![Platform](https://img.shields.io/badge/Platform-Ubuntu%20%2F%20Debian%20Linux-black.svg?style=for-the-badge&logo=linux)]()
[![Memory Footprint](https://img.shields.io/badge/RAM-~30%20MB-black.svg?style=for-the-badge)]()
[![Design System](https://img.shields.io/badge/Design-Monochrome%20B%26W-black.svg?style=for-the-badge)]()
[![License: MIT](https://img.shields.io/badge/License-MIT-black.svg?style=for-the-badge)](LICENSE)

**SparkWA** is an ultra-lightweight, production-grade, headless bridge that connects a dedicated **WhatsApp** account to your existing **Google Gemini Spark** agent through Google's documented Gmail-triggered task schedules.

Designed specifically for headless Ubuntu VPS deployments and accessible to non-technical operators, it features an interactive **Black & White Terminal User Interface (TUI)**, a **Step-by-Step Setup Wizard**, and runs as a single native Linux binary under `systemd` with zero browser, desktop, Node.js, Python, or container dependencies.

---

## Interactive Control Dashboard (TUI)

Non-technical users do not need to memorize terminal commands or arguments:

```bash
# Simply launch sparkwa to open the interactive Control Dashboard
sparkwa

# Or run the step-by-step Guided Setup Wizard directly
sparkwa wizard
```

### Dashboard View

```text
┌─ SYSTEM HEALTH & INTEGRATION SNAPSHOT ─────────────────────────────────────────┐
│  WhatsApp Bot (Number 1)    : +919510678362              [● CONNECTED]         │
│  Authorized User (Number 2) : +918530483089              [● ALLOWLISTED]       │
│  Gmail Bridge Mailbox       : aidrago0021@gmail.com      [● CONNECTED]         │
│  Spark Agent Destination    : ukideashare0021@gmail.com  [● FULLY_VERIFIED]    │
│  SQLite Persistent Queue    : 0 task(s) active in queue  [● HEALTHY]           │
└────────────────────────────────────────────────────────────────────────────────┘

┌─ CONTROL MENU — SELECT OPTION [0-8] ───────────────────────────────────────────┐
│   [1] Guided Setup Wizard   Walk through phone numbers, Gmail & WhatsApp QR   │
│   [2] Pair WhatsApp         Display terminal QR code to link Phone Number 1   │
│   [3] Connect Gmail API     Perform Google OAuth 2.0 browser authorization    │
│   [4] System Diagnostics    Run 6-point diagnostics on config, permissions    │
│   [5] Queue & Status        Inspect bridge capability report & recent tasks   │
│   [6] Send Test Ping        Dispatch safe read-only test ping to Spark        │
│   [7] Start Bridge Daemon   Launch production background listener service     │
│   [8] Documentation & Help  Show WhatsApp slash commands and operational guide│
│                                                                               │
│   [0] Exit                  Exit to command shell                             │
└────────────────────────────────────────────────────────────────────────────────┘
```

---

## Logo Psychology & Monochromatic Design

The visual design of SparkWA is guided by ergonomics and cognitive psychology:

* **The Dual Archetype (The Magician & The Messenger)**:
  * **Gemini Spark**: The spark of intelligence represents transformative thought, swift computation, and illumination.
  * **WhatsApp (Dialogue & Bridge)**: The conversational bridge represents interpersonal communication, trust, and ubiquitous reach.
* **Strict Black & White (Zero-Color ANSI)**:
  * **Anxiety Mitigation**: Traditional terminal red/yellow warning banners trigger panic in non-technical users. A high-contrast monochrome palette eliminates warning fatigue and creates a calm, predictable operating atmosphere.
  * **Swiss International Typographic Authority**: Clean structural frames (`┌─┐`, `│ │`, `└─┘`), high-contrast inverted badges (`\033[7m`), and bold white headings (`\033[1m`) provide the clarity of precision instruments.
  * **Maximal Contrast (21:1)**: Pure Black and White guarantees optimal legibility across any terminal client, light or dark mode, and high-DPI displays.

---

## Architecture

```text
WhatsApp User (Number 2)
       │ (Encrypted WhatsApp Text)
       ▼
WhatsApp Bot Account (Number 1)
       │ (whatsmeow Multi-Device)
       ▼
┌─────────────────────────────────────────────────────────────┐
│                          SPARKWA                            │
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

- **Non-Tech Friendly TUI:** Interactive Black & White Control Dashboard with step-by-step Setup Wizard (`sparkwa wizard`).
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

SparkWA enforces explicit capability states to maintain transparency regarding what is verified versus unverified:

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
scp sparkwa-v1.0.0-linux-amd64.tar.gz credentials.json user@your-vps-ip:/tmp/
```

### 2. Extract and Run Automated Installer
```bash
ssh user@your-vps-ip
cd /tmp && tar -xzf sparkwa-v1.0.0-linux-amd64.tar.gz
cd sparkwa && sudo ./scripts/install.sh
```

### 3. Move Credentials & Run Guided Setup
```bash
sudo mv /tmp/credentials.json /etc/sparkwa/credentials.json
sudo chmod 0600 /etc/sparkwa/credentials.json

# Launch the interactive wizard
sudo -u sparkwa sparkwa wizard --config /etc/sparkwa/config.env
```

### 4. Authenticate Gmail and Pair WhatsApp
You can run these directly from the TUI menu, or via CLI:
```bash
# Authorize Gmail API (via loopback or SSH tunnel port 8085)
sudo -u sparkwa sparkwa gmail-auth --config /etc/sparkwa/config.env

# Pair WhatsApp Number 1 via Terminal QR Code
sudo -u sparkwa sparkwa login --config /etc/sparkwa/config.env
```

### 5. Validate Health and Start Service
```bash
# Run system doctor
sudo -u sparkwa sparkwa doctor --config /etc/sparkwa/config.env

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
