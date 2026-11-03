# Google Gemini Spark Integration Guide

This guide details the exact procedure for connecting your existing **Google Gemini Spark** agent to **SparkWA** using Google's documented **Gmail-triggered Task Schedules**.

---

## 1. Architectural Overview

```
WhatsApp User (Number 2)
       │
       ▼
SparkWA (VPS)
       │
       ▼ (Gmail API)
Bridge Mailbox (BRIDGE_GMAIL)
       │
       ▼ [SPARKWA:REQ_ID] Email Command
Spark-Connected Gmail Inbox (SPARK_GMAIL)
       │
       ▼ (Event Monitor Trigger)
Gemini Spark Agent (Autonomous Execution)
       │
       ▼ Output Channel (Email Reply or Gmail Draft)
SparkWA Gmail Poller
       │
       ▼ WhatsApp Message
WhatsApp User (Number 2)
```

---

## 2. Prerequisites in Google Account

1. **Google Account with Spark Agent**:
   - The Google Account (`SPARK_GMAIL_ADDRESS`) must have access to Gemini Spark / Google AI Workspace features.
   - Ensure **Personal Intelligence** and **Gemini Apps Activity** are enabled in your Google Account settings.
2. **Connected Apps Enabled**:
   - Open [Gemini](https://gemini.google.com).
   - Click **Settings** (gear icon) > **Extensions / Connected Apps**.
   - Enable **Google Workspace** (grants access to Gmail, Google Drive, Docs, and Calendar).

---

## 3. Creating the Gmail-Triggered Spark Schedule

You configure Spark schedules conversationally or through Gemini's task schedule UI.

### Step-by-Step Configuration:

1. Open **Gemini** on your desktop or mobile browser.
2. In the prompt box, type the following schedule creation instruction:

```text
Create a schedule that monitors my Gmail inbox.
Whenever an email arrives from "BRIDGE_GMAIL_ADDRESS" with a subject containing "[SPARKWA:", execute the command contained inside the email body following the SparkWA protocol.
```
*(Replace `BRIDGE_GMAIL_ADDRESS` with your actual bridge email address, e.g. `my-bridge-bot@gmail.com`).*

3. Gemini will present a schedule confirmation card showing:
   - **Trigger**: Incoming email matching `from:BRIDGE_GMAIL_ADDRESS subject:[SPARKWA:`
   - **Frequency**: Automatic (Event-based Gmail monitor)
   - **Task Action**: Execute instructions and format output.

4. Click **Confirm** / **Save Schedule**.

---

## 4. Master Spark Task Prompt (Copy-Paste)

Attach this prompt to your Gemini Spark custom instructions, Gem, or schedule task description. This ensures Spark operates 24/7 autonomously, never pauses to ask for email confirmation, never echoes command envelopes, and always delivers direct results:

```text
You are an autonomous execution worker integrated with SparkWA. You run 24/7 in the background without requiring manual human supervision.

You receive automated tasks via incoming emails with subjects formatted as:
[SPARKWA:REQUEST_ID] <Short Title>

The email body will follow this format:
BEGIN SPARKWA COMMAND
Request ID: <REQUEST_ID>
Authorized sender: <MASKED_PHONE>
Timestamp: <ISO8601>
Command: <USER_INSTRUCTION>
END SPARKWA COMMAND

OPERATING PROTOCOL:
1. IMMEDIATE AUTONOMOUS EXECUTION (24/7):
   - You have standing, permanent authorization to execute tasks and send reply emails automatically.
   - NEVER ask "Should I send this email?", "Do you want me to reply?", or "Shall I proceed?".
   - NEVER pause or wait for user confirmation to send the email response. Send the email reply directly and immediately to the sender.

2. NEVER ECHO OR QUOTE THE INPUT:
   - NEVER repeat, quote, or echo the "BEGIN SPARKWA COMMAND" block in your reply.
   - Extract only the text after "Command: " and treat it as the real instruction to execute.
   - If the command is a conversational greeting (e.g., "Hi", "Hello"), reply directly and politely.
   - If the command is a query, answer it thoroughly.
   - If the command requires Google Workspace tools (Gmail, Drive, Docs, Sheets, Calendar), use them directly to complete the task.

3. STRICT RESPONSE FORMAT:
   Always reply directly to the email thread using this exact structure:

BEGIN SPARKWA RESPONSE
Request ID: <REQUEST_ID>
Status: SUCCESS
Result: <YOUR DIRECT ANSWER, SUMMARY, OR OUTPUT>
END SPARKWA RESPONSE

   - Replace <REQUEST_ID> with the exact Request ID from the incoming email.
   - Put your actual answer or output after "Result: ".
   - Do not include extra conversational text outside of this block.

4. SAFETY GUARDRAILS (DESTRUCTIVE ACTIONS ONLY):
   - Only if a command asks for an irreversible or high-risk action (such as deleting files, mass-emailing clients, or financial transactions), do NOT execute it immediately.
   - Instead, respond with:

BEGIN SPARKWA RESPONSE
Request ID: <REQUEST_ID>
Status: APPROVAL_REQUIRED
Result: This action requires your confirmation: [brief description]. Reply YES to confirm.
END SPARKWA RESPONSE
```

---

## 5. Output Paths & Feasibility Gate

Google Gemini Spark operates in the background, but its return mechanisms depend on your Google Workspace settings. SparkWA supports two verified return channels:

### Output Path A: Direct Email Reply (Recommended)
- Spark generates a reply email addressed to `BRIDGE_GMAIL_ADDRESS` containing `BEGIN SPARKWA RESPONSE`.
- SparkWA's poller discovers this email via Gmail API within 45 seconds and forwards it to WhatsApp.

### Output Path B: Identifiable Draft
- If your Gemini account policy restricts autonomous external outbound emails, instruct Spark to create a **Draft** in the Gmail account.
- Set `SPARK_RESPONSE_MODE="draft_or_email"` in `/etc/sparkwa/config.env`.
- SparkWA checks for drafts matching `[SPARKWA:REQ_ID]` and reads the result.

### Output Path C: Submission-Only Fallback
- If your Gemini Spark setup executes tasks directly in Google Docs, Google Sheets, or Drive without creating emails or drafts, configure:
  ```env
  SPARK_RESPONSE_MODE="submission_only"
  ```
- In this mode, SparkWA acknowledges task submission on WhatsApp and tracks the task as `SUBMISSION_ONLY`. You inspect results directly in Google Workspace.

---

## 6. Read-Only Verification Test

Before enabling destructive or complex tasks, perform this dedicated test:

1. On your VPS, run the diagnostic test command:
   ```bash
   sudo -u sparkwa sparkwa test-gmail --config /etc/sparkwa/config.env
   ```
2. Verify in your Spark Gmail inbox that an email was received with Subject:
   `[SPARKWA:test-...] SparkWA Diagnostics Ping`
3. Observe if Gemini Spark triggers and generates a response.
4. Send a WhatsApp message from Number 2 to Number 1:
   ```text
   /status
   ```
   Confirm that the capability state reports `SPARK_TRIGGER_CONFIGURED` or `RESPONSE_PATH_VERIFIED`.

---

## 7. Troubleshooting Spark Triggers

| Issue | Likely Cause | Resolution |
|-------|--------------|------------|
| Spark schedule does not trigger | Filter mismatch | Ensure the email Subject contains `[SPARKWA:` with no typos in `BRIDGE_GMAIL_ADDRESS`. |
| Spark triggers with long delay | Background scheduling queue | Google Workspace monitors poll on Google's internal schedule. Delays of 1–3 minutes can occur under high Google Cloud load. |
| Spark does not send email reply | Output permissions restricted | Instruct Spark to create a Gmail draft instead, and set `SPARK_RESPONSE_MODE="draft_or_email"`. |
| Task shows APPROVAL_REQUIRED | Guardrail triggered | Reply on WhatsApp with your confirmation (e.g., "Yes, proceed with task <ID>"). |
