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
║  Pure Go • Single Native Binary • Zero Dependencies • By Krish Ujeniya         ║
╚════════════════════════════════════════════════════════════════════════════════╝
```

# SparkWA: WhatsApp Bridge for Google Gemini Spark

**The Headless WhatsApp to Google Gemini Spark Bridge**  
*By Krish Ujeniya*

[![Go Version](https://img.shields.io/badge/Go-1.23%2B-black.svg?style=for-the-badge&logo=go)](https://golang.org)
[![Documentation](https://img.shields.io/badge/Docs-krishujeniya.github.io%2Fsparkwa-black.svg?style=for-the-badge)](https://krishujeniya.github.io/sparkwa/)
[![Platform](https://img.shields.io/badge/Platform-Ubuntu%20%2F%20Debian%20%2F%20macOS-black.svg?style=for-the-badge&logo=linux)]()
[![Memory Usage](https://img.shields.io/badge/RAM-~30%20MB-black.svg?style=for-the-badge)]()
[![Interface](https://img.shields.io/badge/Interface-Black%20%26%20White%20TUI-black.svg?style=for-the-badge)]()
[![License: MIT](https://img.shields.io/badge/License-MIT-black.svg?style=for-the-badge)](LICENSE)

**SparkWA** connects a dedicated WhatsApp phone number to your Google Gemini Spark agent through Gmail. 

You send instructions from your personal WhatsApp chat, SparkWA forwards them to your Gemini Spark agent, and Spark's reply is sent straight back to your phone.

It runs on an Ubuntu or Debian Linux server (or macOS) as a single native program. It uses about 30 MB of memory, runs in the background 24/7, and needs no browser, Docker, Python, or Node.js.

---

## Table of Contents

- [What is SparkWA?](#what-is-sparkwa)
- [Fast 1-Line Installation (Recommended)](#fast-1-line-installation-recommended)
- [Alternative Installation Options](#alternative-installation-options)
- [How It Works](#how-it-works)
- [Interactive Control Dashboard](#interactive-control-dashboard)
- [Quick Start Guide (5 Minutes)](#quick-start-guide-5-minutes)
- [WhatsApp Commands](#whatsapp-commands)
- [Example Prompts for Gemini Spark](#example-prompts-for-gemini-spark)
- [System Requirements](#system-requirements)
- [Security and Privacy](#security-and-privacy)
- [Frequently Asked Questions (FAQ)](#frequently-asked-questions-faq)
- [Documentation Index](#documentation-index)
- [Author](#author)
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

## Fast 1-Line Installation (Recommended)

Run this single command on your Ubuntu, Debian, or Linux server. It automatically installs SparkWA and immediately launches the interactive dashboard:

```bash
curl -fsSL https://raw.githubusercontent.com/krishujeniya/sparkwa/main/scripts/setup-apt.sh | sudo bash
```

Once installed, you can reopen the control dashboard anytime simply by typing:

```bash
sparkwa
```

---

## Alternative Installation Options

### Option 1: Debian Package (`apt`)

```bash
# Download and install the latest official .deb package
curl -LO https://github.com/krishujeniya/sparkwa/releases/latest/download/sparkwa_1.0.0_amd64.deb
sudo apt install ./sparkwa_1.0.0_amd64.deb
```
*For ARM64 servers (e.g. Raspberry Pi, AWS Graviton), substitute `sparkwa_1.0.0_arm64.deb`.*

### Option 2: macOS / Linux (`brew`)

Install via Homebrew:

```bash
brew tap krishujeniya/sparkwa
brew install sparkwa
```

### Option 3: Standalone Binary Release (`tar.gz`)

```bash
curl -LO https://github.com/krishujeniya/sparkwa/releases/latest/download/sparkwa-v1.0.0-linux-amd64.tar.gz
tar -xzf sparkwa-v1.0.0-linux-amd64.tar.gz
cd sparkwa && sudo ./scripts/install.sh
```

### Option 4: Build from Source (Go 1.24+)

```bash
git clone https://github.com/krishujeniya/sparkwa.git
cd sparkwa
make build
sudo ./scripts/install.sh
```

---

## How It Works

SparkWA links two phone numbers and two Gmail accounts:

| Component | Role | Description |
|-----------|------|-------------|
| **Phone Number 1** | Bot Account | A dedicated phone number paired with SparkWA on your server. |
| **Phone Number 2** | Your Phone | Your personal WhatsApp number. Only this number can send instructions. |
| **Bridge Gmail** | Postman | A Gmail mailbox used by SparkWA to send task emails. |
| **Spark Gmail** | Agent Account | The Google account where your Gemini Spark agent runs. |

### 3D Isometric Architecture Diagram

```text
      .─────────────────────────────────────────────────────────────.
     /                                                             /│
    ┌─────────────────────────────────────────────────────────────┐ │
    │                NODE [2] : AUTHORIZED SENDER                 │ │
    │              Your Personal WhatsApp (Number 2)              │ │
    └─────────────────────────────────────────────────────────────┘/
                                   │
                                   │ (1) Send text message or command
                                   ▼
      .─────────────────────────────────────────────────────────────.
     /                                                             /│
    ┌─────────────────────────────────────────────────────────────┐ │
    │                  NODE [1] : WHATSAPP BOT                    │ │
    │             Dedicated WhatsApp SIM (Number 1)               │ │
    └─────────────────────────────────────────────────────────────┘/
                                   │
                                   │ (2) Read by SparkWA on your server
                                   ▼
      .─────────────────────────────────────────────────────────────.
     /                                                             /│
    ┌─────────────────────────────────────────────────────────────┐ │
    │               NODE [E] : SPARKWA LOCAL ENGINE               │ │
    ├─────────────────────────────────────────────────────────────┤ │
    │  • Verifies sender matches Number 2 (silent drop strangers) │ │
    │  • Saves task to local SQLite database (/var/lib/sparkwa)   │ │
    │  • Formats email protocol RFC-2822: [SPARKWA:TASK_ID]       │ │
    │  • Sends email via official Google Gmail API (OAuth 2.0)    │ │
    └─────────────────────────────────────────────────────────────┘/
                                   │
                                   │ (3) Outbound email via Gmail API
                                   ▼
      .─────────────────────────────────────────────────────────────.
     /                                                             /│
    ┌─────────────────────────────────────────────────────────────┐ │
    │          NODE [S] : GOOGLE WORKSPACE & GEMINI SPARK         │ │
    ├─────────────────────────────────────────────────────────────┤ │
    │   Bridge Gmail ─────────────────────► Spark Gmail Inbox     │ │
    │                                             │               │ │
    │                                             │ (4) Trigger   │ │
    │                                             ▼               │ │
    │                                    Gemini Spark Agent       │ │
    │                                    (Drive, Docs, Calendar)  │ │
    │                                             │               │ │
    │                                             │ (5) Reply     │ │
    │   Bridge Gmail ◄────────────────────────────┘               │ │
    └─────────────────────────────────────────────────────────────┘/
                                   │
                                   │ (6) Discovered by SparkWA poller
                                   ▼
      .─────────────────────────────────────────────────────────────.
     /                                                             /│
    ┌─────────────────────────────────────────────────────────────┐ │
    │                NODE [2] : AUTHORIZED SENDER                 │ │
    │              Spark's answer delivered to chat               │ │
    └─────────────────────────────────────────────────────────────┘/
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
┌─ SYSTEM INTEGRATION PIPELINE [PRIVACY-SAFE] ───────────────────────────────────┐
│  Node [1] WhatsApp Bot      : Connected & Listening      [● ONLINE]            │
│  Node [2] Authorized User   : Allowlisted Sender         [● ALLOWED]           │
│  Node [E] Gmail Bridge      : Authenticated (OAuth 2.0)  [● READY]             │
│  Node [S] Gemini Spark      : Integration Verified       [● ACTIVE]            │
│                                                                                │
│  PIPELINE FLOW              : [2 ➔ 1] ➔ [1 ➔ E] ➔ [E ➔ S] ➔ [S ➔ E] ➔ [E ➔ 2] │
│  Persistent Queue           : 0 active task(s) in queue  [● HEALTHY]           │
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

Once installed, follow these steps to pair your WhatsApp and Gmail accounts:

### Step 1: Place Your Google OAuth Credentials
Place your `credentials.json` file from Google Cloud Console into `/etc/sparkwa`:
```bash
sudo mv credentials.json /etc/sparkwa/credentials.json
sudo chmod 0600 /etc/sparkwa/credentials.json
```

### Step 2: Open the Interactive Setup Wizard
Run the command-line wizard:
```bash
sudo sparkwa
```
Select **[1] Guided Setup Wizard**. It will prompt you for:
1. Phone Number 1 (your bot account number in international format, e.g. `+1234567890`).
2. Phone Number 2 (your personal number in international format, e.g. `+1987654321`).
3. Your Bridge Gmail address.
4. Your Gemini Spark Gmail address.

### Step 3: Link WhatsApp & Connect Gmail
From the same interactive dashboard:
- Select **[2] Pair WhatsApp**: Scan the terminal ASCII QR code using Phone Number 1 (**Settings > Linked Devices > Link a Device**).
- Select **[3] Connect Gmail API**: Authorize Google OAuth via browser loopback.
- Select **[4] System Diagnostics**: Run a 6-point check to ensure all connections are healthy.

### Step 4: Start the Background Service
Enable and start the background daemon to run 24/7:
```bash
sudo systemctl enable --now sparkwa
```

Check live status anytime:
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

## Example Prompts for Gemini Spark

Any text message sent from your personal phone (Number 2) that does not start with a slash command (`/`) is treated as an autonomous instruction for Gemini Spark. Spark accesses your Google Workspace tools directly and returns the response.

### 1. Calendar & Scheduling
- **Create an Event**:
  > *"Schedule a 30-minute sync with Alex tomorrow at 3:00 PM and invite alex@company.com"*
  >
  > **Spark Action**: Checks Google Calendar for conflicts, creates the event with a Google Meet link, sends invitations, and reports confirmation back to WhatsApp.
- **Daily Agenda Briefing**:
  > *"What meetings and tasks do I have scheduled for today and tomorrow morning?"*
  >
  > **Spark Action**: Queries Google Calendar and Google Tasks, formats a chronological agenda, and delivers it to WhatsApp.
- **Rescheduling**:
  > *"Move my 2:00 PM team review today to Thursday at 11:00 AM."*
  >
  > **Spark Action**: Locates the calendar entry, checks attendee availability, updates the event, and notifies participants.
- **Reminders & Deadlines**:
  > *"Remind me on Friday at 5:00 PM to submit monthly expenses and invoice reports."*
  >
  > **Spark Action**: Creates a scheduled task with a notification trigger in Google Tasks.

### 2. Gmail & Communications
- **Morning Inbox Summary**:
  > *"Summarize my unread emails from the past 24 hours. Highlight any urgent emails from clients or management."*
  >
  > **Spark Action**: Scans your Gmail inbox, generates concise bullet summaries, and highlights items requiring action.
- **Assisted Drafts**:
  > *"Draft a polite reply to Sarah's email regarding the product roadmap delay, confirming we will ship on Friday."*
  >
  > **Spark Action**: Creates a draft in your Gmail account and sends the preview back to WhatsApp for your review.
- **Search & Lookup**:
  > *"Did invoice #4081 from Acme Corp arrive this week? Check attachments for a PDF."*
  >
  > **Spark Action**: Searches messages matching the query, verifies the PDF attachment, and extracts the invoice total.

### 3. Google Drive, Docs & Sheets
- **Document Summarization**:
  > *"Find the latest 'Q4 Financial Review' presentation in my Google Drive and summarize the key revenue milestones."*
  >
  > **Spark Action**: Locates the presentation, analyzes slides and notes, and delivers key metrics to WhatsApp.
- **Google Sheets Data Entry**:
  > *"Add a row to the '2026 Expense Log' spreadsheet: Nov 3, Client Lunch, $64.20, Paid via Visa."*
  >
  > **Spark Action**: Opens the Google Sheet, appends the row under the correct columns, and confirms the updated sheet balance.
- **Document Creation**:
  > *"Create a new Google Doc named 'Weekly Standup Notes - Nov 3' with sections for Wins, Blockers, and Action Items."*
  >
  > **Spark Action**: Creates the formatted document in Google Drive and returns the direct edit link.

### 4. Safety Guardrails & Approval Confirmation
When an instruction has external impact or irreversible consequences, Gemini Spark halts execution and requests your explicit approval via WhatsApp before proceeding:
- **Mass Email / Bulk Operations**:
  > *"Draft an announcement email to all team leads about Friday server maintenance."*
  >
  > **WhatsApp Approval Flow**:
  > ```text
  > [APPROVAL REQUIRED]
  > Task: 256c0c48
  > Action: Send announcement email to team-all@company.com (18 recipients).
  > 
  > Reply with your confirmation or cancellation.
  > ```
- **File Archival / Deletion**:
  > *"Move all files in the 'Old Scans' folder older than 1 year to Archive."*
  >
  > **Spark Action**: Counts matching files and requests confirmation with the exact count before moving any files.

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

- **[Live Single-Page Documentation Site](https://krishujeniya.github.io/sparkwa/)** — Complete interactive guide with one-click copy commands and system prompt.
- [docs/SPARK_SETUP.md](docs/SPARK_SETUP.md) — Step-by-step instructions to configure your Gemini Spark schedule and prompt.
- [docs/GMAIL_OAUTH_SETUP.md](docs/GMAIL_OAUTH_SETUP.md) — Creating Google Cloud OAuth credentials and connecting Gmail.
- [docs/VPS_INSTALL.md](docs/VPS_INSTALL.md) — Full manual installation and deployment guide for Ubuntu servers.
- [docs/TROUBLESHOOTING.md](docs/TROUBLESHOOTING.md) — Quick solutions for connection, pairing, or token issues.
- [docs/SECURITY.md](docs/SECURITY.md) — Security model, permission settings, and guardrail details.

---

## Author

Created and maintained by **Krish Ujeniya** ([@krishujeniya](https://github.com/krishujeniya)).

---

## License

This project is open-source software licensed under the [MIT License](LICENSE).
