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

# SparkWA: WhatsApp Bridge for Google Gemini Spark

[![Go Version](https://img.shields.io/badge/Go-1.23%2B-black.svg?style=for-the-badge&logo=go)](https://golang.org)
[![Platform](https://img.shields.io/badge/Platform-Ubuntu%20%2F%20Debian%20Linux-black.svg?style=for-the-badge&logo=linux)]()
[![Memory Usage](https://img.shields.io/badge/RAM-~30%20MB-black.svg?style=for-the-badge)]()
[![Interface](https://img.shields.io/badge/Interface-Black%20%26%20White%20TUI-black.svg?style=for-the-badge)]()
[![License: MIT](https://img.shields.io/badge/License-MIT-black.svg?style=for-the-badge)](LICENSE)

**SparkWA** connects a dedicated WhatsApp phone number to your Google Gemini Spark agent through Gmail. 

You send instructions from your personal WhatsApp chat, SparkWA forwards them to your Gemini Spark agent, and Spark's reply is sent straight back to your phone.

It runs on an Ubuntu or Debian Linux server as a single native program. It uses about 30 MB of memory, runs in the background 24/7, and needs no browser, Docker, Python, or Node.js.

---

## Table of Contents

- [What is SparkWA?](#what-is-sparkwa)
- [How It Works](#how-it-works)
- [Interactive Control Dashboard](#interactive-control-dashboard)
- [Quick Start Guide (5 Minutes)](#quick-start-guide-5-minutes)
- [WhatsApp Commands](#whatsapp-commands)
- [System Requirements](#system-requirements)
- [Security and Privacy](#security-and-privacy)
- [Frequently Asked Questions (FAQ)](#frequently-asked-questions-faq)
- [Documentation Index](#documentation-index)
- [License](#license)

---

## What is SparkWA?

SparkWA is a private bridge between WhatsApp and your personal Google Gemini Spark agent.

Standard chatbot tools replace your agent with a basic model API that cannot access your personal files or schedules. In contrast, SparkWA connects to your **actual Gemini Spark agent** inside Google Workspace. This means Spark can access your Google Drive, Docs, Sheets, Calendar, and Gmail while you control it entirely from WhatsApp.

### Key Benefits

- **Simple for Non-Technical Users:** Includes a guided setup wizard and an interactive menu. You do not need to memorize commands.
- **Single File Installation:** Built in Go as one executable file with zero external software requirements.
- **Low Server Cost:** Uses about 30 MB of RAM and 0.0% idle CPU. Runs on basic 512 MB or 1 GB Linux virtual private servers (VPS).
- **Private and Restricted:** Only your personal phone number can send commands. Messages from unknown numbers or group chats are dropped.
- **Safety First:** If Gemini Spark needs confirmation before taking an action (like deleting a file), it forwards the confirmation prompt to your WhatsApp chat before proceeding.

---

## How It Works

SparkWA links two phone numbers and two Gmail accounts:

| Component | Role | Description |
|-----------|------|-------------|
| **Phone Number 1** | Bot Account | A dedicated phone number paired with SparkWA on your server. |
| **Phone Number 2** | Your Phone | Your personal WhatsApp number. Only this number can send instructions. |
| **Bridge Gmail** | Postman | A Gmail mailbox used by SparkWA to send task emails. |
| **Spark Gmail** | Agent Account | The Google account where your Gemini Spark agent runs. |

### Message Flow Diagram

```text
Your Personal WhatsApp (Number 2)
       │
       │ (1) Send text message or command
       ▼
Dedicated WhatsApp Bot (Number 1)
       │
       │ (2) Read by SparkWA on your server
       ▼
┌─────────────────────────────────────────────────────────────┐
│                          SPARKWA                            │
│  • Verifies sender matches Number 2                         │
│  • Saves task to local SQLite database                      │
│  • Formats email: [SPARKWA:TASK_ID]                         │
│  • Sends email via official Google Gmail API                │
└──────────────────────────────┬──────────────────────────────┘
                               │
                               │ (3) Outbound email via Gmail API
                               ▼
┌─────────────────────────────────────────────────────────────┐
│                      GOOGLE WORKSPACE                       │
│  Bridge Gmail ──────────► Spark Gmail Inbox                 │
│                                  │                          │
│                                  │ (4) Event trigger        │
│                                  ▼                          │
│                         Gemini Spark Agent                  │
│                         (Runs task using Google tools)      │
│                                  │                          │
│                                  │ (5) Email reply or draft │
│  Bridge Gmail ◄──────────────────┘                          │
└──────────────────────────────┬──────────────────────────────┘
                               │
                               │ (6) Discovered by SparkWA poller
                               ▼
Your Personal WhatsApp (Number 2) ◄─── Spark's answer delivered to chat
```

---

## Interactive Control Dashboard

SparkWA includes a high-contrast Black and White terminal interface designed for clear reading and easy navigation.

To open the main dashboard:

```bash
sparkwa
```

### Dashboard Screen

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

You can choose any action by pressing a single number key (`1` through `8`), or navigate with the arrow keys and press Enter.

---

## Quick Start Guide (5 Minutes)

Follow these five steps to install SparkWA on a clean Ubuntu or Debian server.

### Step 1: Upload Files to Your Server

Copy the downloaded release archive and your Google `credentials.json` file to your server:

```bash
scp sparkwa-v1.0.0-linux-amd64.tar.gz credentials.json user@your-server-ip:/tmp/
```

### Step 2: Run the Installer

Log in to your server and run the installation script:

```bash
ssh user@your-server-ip
cd /tmp && tar -xzf sparkwa-v1.0.0-linux-amd64.tar.gz
cd sparkwa && sudo ./scripts/install.sh
```

The installer places the program at `/usr/local/bin/sparkwa`, creates a dedicated system user, and configures the background service.

### Step 3: Run the Setup Wizard

Move your Google credentials into place and start the guided wizard:

```bash
sudo mv /tmp/credentials.json /etc/sparkwa/credentials.json
sudo chmod 0600 /etc/sparkwa/credentials.json

# Launch the wizard
sudo -u sparkwa sparkwa wizard --config /etc/sparkwa/config.env
```

The wizard will prompt you for:
1. Phone Number 1 (your bot account number in international format, e.g. `+1234567890`).
2. Phone Number 2 (your personal number in international format, e.g. `+1987654321`).
3. Your Bridge Gmail address.
4. Your Gemini Spark Gmail address.

### Step 4: Link Gmail and WhatsApp

Run these two commands to complete authorization:

```bash
# 1. Authorize Gmail API access
sudo -u sparkwa sparkwa gmail-auth --config /etc/sparkwa/config.env

# 2. Display the WhatsApp pairing QR code in your terminal
sudo -u sparkwa sparkwa login --config /etc/sparkwa/config.env
```

Open WhatsApp on phone Number 1, go to **Settings > Linked Devices > Link a Device**, and scan the QR code displayed in your terminal.

### Step 5: Start the Background Service

Start the service and enable it to run automatically on system boot:

```bash
sudo systemctl start sparkwa
sudo systemctl enable sparkwa
```

Check the live status anytime:

```bash
sudo systemctl status sparkwa
```

---

## WhatsApp Commands

Send these commands from your personal phone (Number 2) to your bot (Number 1):

| Command | Action |
|---------|--------|
| `/ping` | Test if the bot and server are online. |
| `/status` | View health details for WhatsApp, Gmail, and active tasks. |
| `/jobs` | List your 5 most recent tasks and their progress. |
| `/job <id>` | View complete details and history for a specific task ID. |
| `/cancel <id>` | Stop a task that is waiting or in progress. |
| `/help` | Print the command manual directly in your chat. |
| **Any other text** | Forwarded to Gemini Spark as an instruction for your agent to complete. |

---

## System Requirements

| Requirement | Minimum Specification | Notes |
|-------------|-----------------------|-------|
| **Operating System** | Ubuntu 22.04 / 24.04 LTS or Debian 12 | Works on 64-bit x86 (`amd64`) and ARM (`arm64`). |
| **Memory (RAM)** | 512 MB | SparkWA process consumes ~30 MB. |
| **Processor (CPU)** | 1 shared vCPU | Idle CPU usage is 0.0%. |
| **Disk Space** | 50 MB | Holds the executable file and SQLite database. |
| **Network** | Outbound HTTPS (port 443) and WhatsApp (port 5222) | No public inbound ports required. |
| **Dependencies** | None | No Docker, Node.js, Python, or web browser needed. |

---

## Security and Privacy

SparkWA is built with strict privacy controls to keep your data safe:

- **Single Sender Allowlist:** Only your verified personal phone number can send instructions. All unknown numbers, spam calls, group messages, and status updates are discarded without a reply.
- **Official Google OAuth 2.0:** Uses official Google APIs with narrow permissions (`gmail.send` and `gmail.readonly`). SparkWA never sees or stores your Google account password.
- **No Inbound Web Ports:** The server does not run an exposed web server. All communication goes out through encrypted HTTPS and WhatsApp connections.
- **Confirmation Safety Guardrail:** When Gemini Spark detects an action with external impact (such as deleting files or sending emails to third parties), SparkWA sends the confirmation question to your WhatsApp chat and waits for your explicit approval.
- **Restricted Server Permissions:** Runs under an unprivileged `sparkwa` service account with read-only system file isolation and private temporary storage.

---

## Frequently Asked Questions (FAQ)

### Can I use my personal WhatsApp number as the bot?
No. You need two numbers:
1. **Phone Number 1 (Bot Account):** A dedicated SIM or eSIM number paired with the server.
2. **Phone Number 2 (Your Phone):** Your regular personal WhatsApp account that you use daily to message the bot.

### Does SparkWA replace my Gemini Spark agent with an API bot?
No. SparkWA connects to your actual, existing Gemini Spark agent in Google Workspace. It preserves your connected tools, Google Drive access, Docs, Sheets, and Calendar schedules.

### Does my server need a desktop or browser installed?
No. SparkWA is a command-line program built specifically for headless Linux servers. It includes a built-in terminal QR code display and a console OAuth helper.

### What happens if the server reboots?
SparkWA runs under `systemd`. It starts up automatically on boot, reconnects to WhatsApp and Gmail, and resumes any pending tasks from its local SQLite database.

### How do I update SparkWA when a new version is released?
Download the new release archive and run the update script:
```bash
sudo ./scripts/update.sh sparkwa-v1.x.x-linux-amd64.tar.gz
```
Your settings, paired WhatsApp session, and task history are saved automatically.

---

## Documentation Index

Detailed guides are located in the [`docs/`](docs/) folder:

- [docs/SPARK_SETUP.md](docs/SPARK_SETUP.md) — Step-by-step instructions to configure your Gemini Spark schedule and prompt.
- [docs/GMAIL_OAUTH_SETUP.md](docs/GMAIL_OAUTH_SETUP.md) — Creating Google Cloud OAuth credentials and connecting Gmail.
- [docs/VPS_INSTALL.md](docs/VPS_INSTALL.md) — Full manual installation and deployment guide for Ubuntu servers.
- [docs/TROUBLESHOOTING.md](docs/TROUBLESHOOTING.md) — Quick solutions for connection, pairing, or token issues.
- [docs/SECURITY.md](docs/SECURITY.md) — Security model, permission settings, and guardrail details.
- [docs/VERIFICATION.md](docs/VERIFICATION.md) — Automated test results and performance measurements.

---

## License

This project is open-source software licensed under the [MIT License](LICENSE).
