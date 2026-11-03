# SparkWA Security Policy & Architecture

Security and operational reliability are primary design pillars of **SparkWA**. This document details the threat model, defensive controls, privilege separation, and disclosures.

---

## 1. Threat Model & Defensive Invariants

| Threat | Impact | Defensive Control in SparkWA |
|--------|--------|------------------------------|
| **Unauthorized WhatsApp Senders** | Command injection, unauthorized AI task execution | **Strict single-user allowlist**: Fail-closed verification matching normalized E.164 phone numbers. Unknown senders and unmapped LIDs are dropped without response. |
| **WhatsApp Group Message Ingestion** | Accidental command execution from shared chats | **Group messages dropped**: Any message where `evt.Info.IsGroup == true` is silently discarded before parsing. |
| **Broadcast / Status Flood** | Resource exhaustion, queue poisoning | **Broadcasts ignored**: Status announcements and broadcast lists are dropped immediately. |
| **Looping Self-Replies** | Infinite recursive command dispatch | **Self-messages dropped**: `evt.Info.IsFromMe == true` messages are never processed as commands. |
| **Arbitrary Shell Execution** | Host compromise | **Zero shell execution**: SparkWA contains no shell evaluation (`exec.Command` or `sh -c`) for chat input. All user messages are routed purely through structured API requests. |
| **Malicious Email Spoofing** | Fake Spark task results injected into WhatsApp | **Strict sender & correlation checks**: Incoming emails must originate from the verified `SPARK_GMAIL_ADDRESS` and must match an active task's unique Request ID. |
| **Credential Theft on VPS** | Google account compromise | **Linux privilege isolation**: Service runs as a dedicated non-root user `sparkwa` with restrictive `0600` file permissions on tokens and `ProtectSystem=strict` systemd sandboxing. |

---

## 2. Principle of Least Privilege

### Dedicated Non-Root User
SparkWA executes under the dedicated Linux system user `sparkwa` (`/usr/sbin/nologin`). It does not run as root.

### Filesystem Permissions
- `/etc/sparkwa/config.env`: Mode `0600`, owned by `sparkwa:sparkwa`.
- `/etc/sparkwa/credentials.json`: Mode `0600`, owned by `root:sparkwa`.
- `/var/lib/sparkwa/token.json`: Mode `0600`, owned by `sparkwa:sparkwa`.
- `/var/lib/sparkwa/*.db`: Mode `0600`, located inside directory mode `0700`.

### Systemd Hardening Flags
The installed service unit enforces Linux kernel sandboxing:
- `ProtectSystem=strict`: The entire OS filesystem is mounted read-only except `/var/lib/sparkwa`.
- `ProtectHome=yes`: `/home`, `/root`, and `/run/user` are completely invisible to the process.
- `NoNewPrivileges=yes`: Prevents privilege escalation.
- `PrivateTmp=yes`: Isolated temporary directories.
- `PrivateDevices=yes`: Access to physical hardware devices is blocked.

### Minimal Gmail Scopes
SparkWA requests only the minimum required Gmail API scopes:
1. `https://www.googleapis.com/auth/gmail.send`: Allows dispatching command emails to the Spark mailbox.
2. `https://www.googleapis.com/auth/gmail.readonly`: Allows reading responses and drafts.
It does **not** request broad account access or delete permissions.

---

## 3. Preserving Gemini Spark Approval Guardrails

Google Gemini Spark includes safety guardrails requiring user confirmation before executing sensitive external actions (e.g., deleting documents, making outbound transactions, sending emails).

**SparkWA will NEVER attempt to bypass, spoof, or auto-confirm an approval prompt.**

When Spark flags an action as requiring approval (either via `Status: APPROVAL_REQUIRED` or phrases like *"Do you want me to proceed with..."*):
1. The task transitions to `APPROVAL_REQUIRED`.
2. SparkWA sends an alert to WhatsApp Number 2 displaying the exact confirmation request.
3. Execution remains paused until you review the request and explicitly reply on WhatsApp.

---

## 4. WhatsApp Automation Disclosures & Risk Management

> [!WARNING]
> **WhatsApp Terms of Service Disclosure:**
> WhatsApp's Terms of Service restrict unauthorized automation on personal consumer accounts. The underlying multi-device library (`whatsmeow`) simulates WhatsApp Web Linked Devices.
> 
> To mitigate account risk:
> 1. **Use a dedicated SIM for Phone Number 1** (the bot account). Never use your primary personal phone number as the bot account.
> 2. SparkWA uses conservative rate limits, delays between multi-part chunks, and low message frequency.
> 3. Senders other than Number 2 are ignored without replies, avoiding spam flags from unknown contacts.
