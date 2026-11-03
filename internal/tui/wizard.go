package tui

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/sparkwa/sparkwa/internal/config"
	"github.com/sparkwa/sparkwa/internal/identity"
)

// RunSetupWizard guides a non-technical user step-by-step through configuring SparkWA.
func RunSetupWizard(ctx context.Context, in io.Reader, out io.Writer, cfg *config.Config) error {
	reader := bufio.NewReader(in)

	fmt.Fprintln(out)
	fmt.Fprintln(out, Bold("================================================================================"))
	fmt.Fprintln(out, Bold("                   SPARKWA GUIDED SETUP WIZARD                                  "))
	fmt.Fprintln(out, Bold("================================================================================"))
	fmt.Fprintln(out, "This wizard will help you configure your WhatsApp phone numbers and Gmail accounts")
	fmt.Fprintln(out, "step-by-step. Press Enter at any prompt to keep the existing value in brackets.")
	fmt.Fprintln(out, "================================================================================")
	fmt.Fprintln(out)

	// Step 1: WhatsApp Bot Account (Number 1)
	fmt.Fprintln(out, Bold("STEP 1 OF 4: WHATSAPP BOT ACCOUNT (PHONE NUMBER 1)"))
	fmt.Fprintln(out, Dim("This is the dedicated WhatsApp account (e.g. on a secondary SIM) that acts as the bot."))
	fmt.Fprintln(out, Dim("Format: international format starting with + and country code (e.g. +1234567890)"))

	currentBot := cfg.BotPhoneNumber
	for {
		prompt := "Enter Bot Phone Number"
		if currentBot != "" {
			prompt = fmt.Sprintf("Enter Bot Phone Number [%s]", currentBot)
		}
		fmt.Fprintf(out, "%s: ", prompt)

		line, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		val := strings.TrimSpace(line)
		if val == "" && currentBot != "" {
			val = currentBot
		}
		normalized, err := identity.NormalizeE164(val)
		if err != nil {
			fmt.Fprintf(out, "  %s %v. Please include country code, e.g. +14155552671.\n", Inverse("[INVALID NUMBER]"), err)
			continue
		}
		cfg.BotPhoneNumber = normalized
		fmt.Fprintf(out, "  ✓ Bot number configured: %s\n\n", Bold(cfg.BotPhoneNumber))
		break
	}

	// Step 2: Authorized Personal Phone Number (Number 2)
	fmt.Fprintln(out, Bold("STEP 2 OF 4: YOUR PERSONAL PHONE NUMBER (PHONE NUMBER 2)"))
	fmt.Fprintln(out, Dim("This is YOUR personal WhatsApp number. SparkWA strictly fails closed and will ONLY"))
	fmt.Fprintln(out, Dim("accept messages from this number. Messages from any other person or group are dropped."))

	currentAuth := cfg.AuthorizedPhoneNumber
	for {
		prompt := "Enter Your Personal Phone Number"
		if currentAuth != "" {
			prompt = fmt.Sprintf("Enter Your Personal Phone Number [%s]", currentAuth)
		}
		fmt.Fprintf(out, "%s: ", prompt)

		line, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		val := strings.TrimSpace(line)
		if val == "" && currentAuth != "" {
			val = currentAuth
		}
		normalized, err := identity.NormalizeE164(val)
		if err != nil {
			fmt.Fprintf(out, "  %s %v. Please include country code, e.g. +14155552671.\n", Inverse("[INVALID NUMBER]"), err)
			continue
		}
		if normalized == cfg.BotPhoneNumber {
			fmt.Fprintf(out, "  %s Bot number and personal number cannot be identical. Use two distinct numbers.\n", Inverse("[CONFLICT]"))
			continue
		}
		cfg.AuthorizedPhoneNumber = normalized
		fmt.Fprintf(out, "  ✓ Authorized personal number configured: %s\n\n", Bold(cfg.AuthorizedPhoneNumber))
		break
	}

	// Step 3: Google Gmail Mailboxes
	fmt.Fprintln(out, Bold("STEP 3 OF 4: GMAIL MAILBOX ADDRESSES"))
	fmt.Fprintln(out, Dim("SparkWA uses two Gmail accounts to communicate with your Google Gemini Spark agent:"))
	fmt.Fprintln(out, Dim("  1. Bridge Gmail Address: The account used by SparkWA on your VPS."))
	fmt.Fprintln(out, Dim("  2. Spark Gmail Address:  The Google account connected to Gemini Spark."))

	currentBridge := cfg.BridgeGmailAddress
	for {
		prompt := "Enter Bridge Gmail Address"
		if currentBridge != "" {
			prompt = fmt.Sprintf("Enter Bridge Gmail Address [%s]", currentBridge)
		}
		fmt.Fprintf(out, "%s: ", prompt)

		line, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		val := strings.TrimSpace(line)
		if val == "" && currentBridge != "" {
			val = currentBridge
		}
		if !strings.Contains(val, "@") || !strings.Contains(val, ".") {
			fmt.Fprintf(out, "  %s Please enter a valid email address.\n", Inverse("[INVALID EMAIL]"))
			continue
		}
		cfg.BridgeGmailAddress = val
		fmt.Fprintf(out, "  ✓ Bridge Gmail address: %s\n\n", Bold(cfg.BridgeGmailAddress))
		break
	}

	currentSpark := cfg.SparkGmailAddress
	for {
		prompt := "Enter Spark-Connected Gmail Address"
		if currentSpark != "" {
			prompt = fmt.Sprintf("Enter Spark-Connected Gmail Address [%s]", currentSpark)
		}
		fmt.Fprintf(out, "%s: ", prompt)

		line, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		val := strings.TrimSpace(line)
		if val == "" && currentSpark != "" {
			val = currentSpark
		}
		if !strings.Contains(val, "@") || !strings.Contains(val, ".") {
			fmt.Fprintf(out, "  %s Please enter a valid email address.\n", Inverse("[INVALID EMAIL]"))
			continue
		}
		cfg.SparkGmailAddress = val
		fmt.Fprintf(out, "  ✓ Spark-connected Gmail address: %s\n\n", Bold(cfg.SparkGmailAddress))
		break
	}

	// Step 4: OAuth Client Credentials Check
	fmt.Fprintln(out, Bold("STEP 4 OF 4: GOOGLE OAUTH CREDENTIALS CHECK"))
	targetCreds := cfg.GmailCredentialsPath
	if targetCreds == "" {
		targetCreds = "credentials.json"
	}

	if _, err := os.Stat(targetCreds); err == nil {
		fmt.Fprintf(out, "  ✓ OAuth credentials file found: %s\n\n", Bold(targetCreds))
	} else {
		fmt.Fprintf(out, "  %s Credentials file not found at %s\n", Dim("[NOTE]"), targetCreds)
		fmt.Fprintln(out, Dim("  To connect Google Gmail, you need credentials.json from Google Cloud Console:"))
		fmt.Fprintln(out, Dim("  1. Visit: https://console.cloud.google.com/apis/credentials"))
		fmt.Fprintln(out, Dim("  2. Create OAuth Client ID (Desktop App) and download the JSON file."))
		fmt.Fprintf(out, Dim("  3. Place it at: %s\n\n"), targetCreds)
	}

	// Save Configuration
	targetConfig := cfg.ConfigPath
	if targetConfig == "" {
		if os.Geteuid() == 0 {
			targetConfig = "/etc/sparkwa/config.env"
		} else {
			targetConfig = "./config.env"
		}
	}

	fmt.Fprintf(out, "Save configuration to %s? [Y/n]: ", Bold(targetConfig))
	confirmLine, _ := reader.ReadString('\n')
	confirm := strings.ToLower(strings.TrimSpace(confirmLine))
	if confirm == "" || confirm == "y" || confirm == "yes" {
		if err := saveConfig(targetConfig, cfg); err != nil {
			fmt.Fprintf(out, "  %s Failed to save config: %v\n", Inverse("[ERROR]"), err)
		} else {
			cfg.ConfigPath = targetConfig
			fmt.Fprintf(out, "  ✓ Configuration saved successfully to: %s\n\n", Bold(targetConfig))
		}
	}

	fmt.Fprintln(out, Bold("================================================================================"))
	fmt.Fprintln(out, Bold("                          WIZARD COMPLETE!                                      "))
	fmt.Fprintln(out, Bold("================================================================================"))
	fmt.Fprintln(out, "Next recommended actions:")
	fmt.Fprintln(out, " • Select [2] from the Control Menu to pair WhatsApp via Terminal QR code.")
	fmt.Fprintln(out, " • Select [3] from the Control Menu to authorize Gmail OAuth.")
	fmt.Fprintln(out, " • Select [4] to run health diagnostics and verify all systems.")
	fmt.Fprintln(out, Bold("================================================================================"))
	fmt.Fprint(out, "\nPress Enter to return to the Control Menu...")
	_, _ = reader.ReadString('\n')

	return nil
}

func saveConfig(path string, cfg *config.Config) error {
	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		_ = os.MkdirAll(dir, 0750)
	}

	content := fmt.Sprintf(`# SparkWA Configuration File
# WhatsApp Settings
BOT_PHONE_NUMBER="%s"         # Number 1 (bot account)
AUTHORIZED_PHONE_NUMBER="%s"  # Number 2 (your personal number, only permitted sender)

# Gmail Settings
BRIDGE_GMAIL_ADDRESS="%s"  # Mailbox used by the bridge
SPARK_GMAIL_ADDRESS="%s"   # Google account connected to Gemini Spark

# Runtime Paths & Polling
DATA_DIR="%s"
GMAIL_CREDENTIALS_FILE="%s"
GMAIL_TOKEN_FILE="%s"
GMAIL_POLL_INTERVAL="%s"
TASK_TIMEOUT="%s"
LOG_LEVEL="%s"
`,
		cfg.BotPhoneNumber,
		cfg.AuthorizedPhoneNumber,
		cfg.BridgeGmailAddress,
		cfg.SparkGmailAddress,
		cfg.DataDir,
		cfg.GmailCredentialsPath,
		cfg.GmailTokenPath,
		cfg.GmailPollInterval.String(),
		cfg.TaskTimeout.String(),
		cfg.LogLevel,
	)

	return os.WriteFile(path, []byte(content), 0600)
}
