# Changelog

All notable changes to **SparkWA** will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2026-10-31

### Added
- **Core Architecture:**
  - Pure Go single-binary implementation with zero browser, Node.js, Python, or container dependencies.
  - Multi-device WhatsApp connectivity powered by `whatsmeow`.
  - Secure offline session persistence with embedded SQLite (`modernc.org/sqlite`).
  - Terminal QR code rendering via `github.com/mdp/qrterminal/v3` for headless SSH VPS pairing.
- **Security & Authorization:**
  - Fail-closed sender verification: only WhatsApp Number 2 can issue commands.
  - Automatic suppression of group messages, broadcast statuses, and self messages.
  - Resolution of WhatsApp Linked Identifiers (LID) mapped to verified E.164 phone numbers.
  - Deduplication of incoming WhatsApp and Gmail message IDs.
- **Gemini Spark Bridge:**
  - Official Google Gmail API integration using least-privilege OAuth 2.0 scopes (`gmail.send`, `gmail.readonly`).
  - Structured request envelopes (`[SPARKWA:REQUEST_ID]`) designed for Gemini Spark Gmail monitors.
  - Bidirectional response discovery supporting both direct email replies and identifiable drafts.
  - Capability state evaluator: `NOT_CONFIGURED`, `GMAIL_CONNECTED`, `SPARK_TRIGGER_CONFIGURED`, `RESPONSE_PATH_VERIFIED`, `SUBMISSION_ONLY`, `FULLY_VERIFIED`.
  - Agent action approval detection with explicit interactive confirmation prompts.
- **CLI Commands:**
  - `sparkwa init`: Prepares filesystem directories and generates secure configuration templates.
  - `sparkwa login`: Renders ASCII QR pairing code in terminal.
  - `sparkwa gmail-auth`: Headless OAuth 2.0 helper supporting local loopback callbacks and SSH port forwarding.
  - `sparkwa doctor`: Comprehensive diagnostic engine inspecting config, permissions, SQLite, WhatsApp, Gmail, and Spark capabilities.
  - `sparkwa run`: Hardened background service daemon.
  - `sparkwa status`: Real-time queue and integration status inspector.
  - `sparkwa test-gmail`: Outbound diagnostic probe to verify Gmail monitor triggers.
- **WhatsApp Slash Commands:**
  - `/help`: Detailed command guide.
  - `/status`: Live bridge, WhatsApp, Gmail, and Spark health report.
  - `/jobs`: Summary of recent tasks.
  - `/job <id>`: Detailed view of a specific task.
  - `/cancel <id>`: Safe task cancellation.
  - `/ping`: Latency and responsiveness check.
- **Deployment & Production:**
  - Hardened systemd service file with full Linux sandboxing (`ProtectSystem=strict`, `NoNewPrivileges=yes`).
  - Automated VPS installer (`scripts/install.sh`) with dedicated `sparkwa` service user.
  - Safe uninstaller, updater, and atomic backup scripts.
  - Cross-compilation support for both `linux/amd64` and `linux/arm64`.
- **Documentation:**
  - Complete configuration guide for Gemini Spark agent Gmail monitors (`docs/SPARK_SETUP.md`).
  - Step-by-step GCP OAuth setup guide (`docs/GMAIL_OAUTH_SETUP.md`).
  - VPS deployment and troubleshooting manuals.
