# SparkWA Troubleshooting Guide

This guide covers common issues, root causes, and resolutions for SparkWA.

---

## 1. Diagnostics First (`sparkwa doctor`)

Whenever something is not behaving as expected, stop the service and run the diagnostic doctor:

```bash
sudo systemctl stop sparkwa
sudo -u sparkwa sparkwa doctor --config /etc/sparkwa/config.env
```

The output will pinpoint which subsystem is failing.

---

## 2. WhatsApp Issues

### Issue A: Terminal QR Code Looks Garbled
- **Cause:** Your SSH terminal client is using a non-UTF8 font or improper character encoding.
- **Resolution:**
  - In your terminal emulator (PuTTY, iTerm2, Windows Terminal, GNOME Terminal), set the character encoding to **UTF-8**.
  - Increase terminal font size or widen the window so half-block ASCII characters align properly.

### Issue B: "Session logged out or revoked"
- **Cause:** WhatsApp Web was manually disconnected from the phone, or WhatsApp rotated session security keys.
- **Resolution:**
  1. Stop the background service:
     ```bash
     sudo systemctl stop sparkwa
     ```
  2. Re-pair the device:
     ```bash
     sudo -u sparkwa sparkwa login --config /etc/sparkwa/config.env
     ```
  3. Scan the new QR code with phone Number 1.
  4. Restart the service:
     ```bash
     sudo systemctl start sparkwa
     ```

### Issue C: Bot Ignores WhatsApp Messages
- **Cause:**
  1. The sender is not Phone Number 2 (`AUTHORIZED_PHONE_NUMBER`).
  2. The message was sent in a group chat (groups are ignored).
  3. The phone number in `config.env` does not match the exact international E.164 format.
- **Resolution:**
  - Verify that `AUTHORIZED_PHONE_NUMBER` in `/etc/sparkwa/config.env` starts with `+` and includes the country code (e.g. `+1987654321`).
  - Run `sudo journalctl -u sparkwa -f` while sending a message. If the sender is unauthorized, a debug log will show that the sender was ignored.

---

## 3. Gmail and OAuth Issues

### Issue A: "failed to refresh token: oauth2: cannot fetch token: 400 Bad Request"
- **Cause:**
  - The Google Cloud OAuth consent screen is in **"Testing"** mode, and 7 days have passed since authorization. In Testing mode, Google expires refresh tokens after 7 days.
- **Resolution:**
  1. Go to [Google Cloud Console > OAuth Consent Screen](https://console.cloud.google.com/apis/credentials/consent).
  2. Click **Publish App** to switch from "Testing" to "In Production".
  3. Re-run authorization on your VPS:
     ```bash
     sudo systemctl stop sparkwa
     sudo -u sparkwa sparkwa gmail-auth --config /etc/sparkwa/config.env
     sudo systemctl start sparkwa
     ```

### Issue B: "authenticated account does not match configured bridge address"
- **Cause:**
  - You logged into Google during `sparkwa gmail-auth` using your personal or Spark account instead of `BRIDGE_GMAIL_ADDRESS`.
- **Resolution:**
  - Run `sparkwa gmail-auth` again and ensure you select the bridge account when granting permissions in your browser.

---

## 4. SQLite Database & Lock Conflicts

### Issue: "database is locked" or "cannot start client while service is running"
- **Cause:**
  - Both the `sparkwa.service` background daemon and a manual CLI command (e.g. `sparkwa login` or `sparkwa run`) are trying to access the same SQLite database files simultaneously.
- **Resolution:**
  - Always stop the background systemd service before running interactive CLI commands that mutate sessions:
    ```bash
    sudo systemctl stop sparkwa
    # perform command
    sudo -u sparkwa sparkwa login
    sudo systemctl start sparkwa
    ```

---

## 5. Gemini Spark Trigger Latency

### Issue: Spark takes 1–3 minutes to process commands
- **Cause:**
  - Google's Gmail Monitor schedules do not execute synchronously on email arrival; Google's backend worker polls inbox filters periodically.
  - SparkWA's default poll interval is 45 seconds to stay well below Gmail API quotas.
- **Expected Timing:**
  - Normal round-trip time between command submission on WhatsApp and final result delivery is between **45 and 180 seconds**.

---

## 6. Systemd Service Failures

Inspect the systemd exit code and full log stream:

```bash
sudo systemctl status sparkwa.service
sudo journalctl -u sparkwa -n 50 --no-pager
```

Common causes:
- File permissions: Ensure `/var/lib/sparkwa` is owned by `sparkwa:sparkwa` (`sudo chown -R sparkwa:sparkwa /var/lib/sparkwa`).
- Missing config: Ensure `/etc/sparkwa/config.env` exists and is readable by the `sparkwa` group (`0600` or `0640`).
