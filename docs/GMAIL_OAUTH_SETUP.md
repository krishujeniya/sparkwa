# Google Gmail OAuth 2.0 Setup Guide

SparkWA Lite uses official Google Gmail APIs with least-privilege OAuth 2.0 scopes. It does **not** use deprecated Out-Of-Band (OOB) authentication, does not store Google account passwords, and never exposes public HTTP listeners.

---

## 1. Google Cloud Console Setup

### Step 1: Create or Select a GCP Project
1. Go to the [Google Cloud Console](https://console.cloud.google.com/).
2. Click the project dropdown at the top and click **New Project**.
3. Name it `SparkWA-Bridge` and click **Create**.

### Step 2: Enable the Gmail API
1. In the search bar at the top, type `Gmail API` and select it from Marketplace.
2. Click **Enable**.

### Step 3: Configure the OAuth Consent Screen
1. In the left navigation menu, go to **APIs & Services** > **OAuth consent screen**.
2. Select **External** (or **Internal** if using Google Workspace organization) and click **Create**.
3. Fill in the required fields:
   - **App name**: `SparkWA Lite`
   - **User support email**: Your personal email.
   - **Developer contact information**: Your personal email.
4. Click **Save and Continue**.
5. On the **Scopes** page, click **Add or Remove Scopes**.
   Add the following minimum scopes:
   - `https://www.googleapis.com/auth/gmail.send` (Send messages on your behalf)
   - `https://www.googleapis.com/auth/gmail.readonly` (Read messages and drafts)
6. Click **Save and Continue**.
7. On the **Test Users** page, click **+ Add Users** and add:
   - Your `BRIDGE_GMAIL_ADDRESS`
   - Your personal email
8. Click **Save and Continue**.

> [!IMPORTANT]
> **Testing Mode Note:** In Google Cloud's "Testing" status, OAuth refresh tokens expire after **7 days** unless your GCP project is moved to "In Production" status. To avoid re-authorizing weekly, click **Publish App** on the OAuth Consent Screen page (even without full verification, published personal apps with personal scopes retain persistent refresh tokens).

### Step 4: Create OAuth 2.0 Client Credentials
1. Go to **APIs & Services** > **Credentials**.
2. Click **+ Create Credentials** > **OAuth client ID**.
3. Select **Application type**:
   - Choose **Desktop app** (recommended for loopback redirect) OR **Web application**.
   - If choosing *Web application*, add Authorized redirect URI:
     `http://127.0.0.1:8085/oauth/callback`
4. Click **Create**.
5. Click **Download JSON** on the confirmation dialog.
6. Rename the downloaded file to `credentials.json`.

---

## 2. Deploying Credentials to VPS

Copy `credentials.json` from your development machine to your VPS:

```bash
scp credentials.json user@your-vps-ip:/tmp/credentials.json
```

On the VPS, move it to the protected configuration directory:

```bash
sudo mkdir -p /etc/sparkwa-lite
sudo mv /tmp/credentials.json /etc/sparkwa-lite/credentials.json
sudo chown root:sparkwa /etc/sparkwa-lite/credentials.json
sudo chmod 0600 /etc/sparkwa-lite/credentials.json
```

---

## 3. Running Headless OAuth Authorization

Because your Ubuntu VPS has no graphical interface or browser, SparkWA Lite provides a secure loopback helper that coordinates with your desktop browser.

### Method 1: SSH Port Forwarding (Recommended)

1. Connect to your VPS with port 8085 forwarded to your local machine:
   ```bash
   ssh -L 8085:127.0.0.1:8085 user@your-vps-ip
   ```

2. Run the authorization command on the VPS:
   ```bash
   sudo -u sparkwa sparkwa gmail-auth --config /etc/sparkwa-lite/config.env
   ```

3. SparkWA Lite will print an authorization link:
   ```text
   Open this authorization URL in your browser:
   https://accounts.google.com/o/oauth2/auth?...
   ```

4. Copy the link and open it in your desktop browser.
5. Log in with your **Bridge Gmail Account** (`BRIDGE_GMAIL_ADDRESS`), review permissions, and click **Allow**.
6. Google will redirect to `http://127.0.0.1:8085/oauth/callback`. Because port 8085 is forwarded via SSH, the request reaches your VPS loopback server automatically!
7. The terminal will display:
   ```text
   ✓ Received authorization code via loopback callback!
   ✓ Successfully saved OAuth token with refresh capabilities to /var/lib/sparkwa-lite/token.json (permissions 0600)
   ✓ Successfully verified Gmail API connection for: bot-bridge@gmail.com
   ```

### Method 2: Manual Terminal Paste Fallback

If you cannot forward ports via SSH:
1. Open the URL printed by `sparkwa gmail-auth` in your desktop browser.
2. After clicking "Allow", your browser will attempt to open `http://127.0.0.1:8085/oauth/callback?state=...&code=4/0A...`.
3. Even though the browser shows "Connection Refused", simply copy the entire URL from the browser's address bar.
4. Paste that URL directly into the terminal prompt:
   ```text
   Auth Code or Redirect URL: http://127.0.0.1:8085/oauth/callback?code=4/0A...
   ```
5. SparkWA Lite extracts the code and exchanges it for the token automatically.

---

## 4. Token Security

- Stored at: `/var/lib/sparkwa-lite/token.json`
- Permissions: `0600` (readable only by the `sparkwa` service account).
- SparkWA Lite automatically uses the `refresh_token` to renew expired access tokens silently in the background.
