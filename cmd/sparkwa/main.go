package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/sparkwa/sparkwa/internal/config"
	"github.com/sparkwa/sparkwa/internal/gmail"
	"github.com/sparkwa/sparkwa/internal/health"
	"github.com/sparkwa/sparkwa/internal/identity"
	"github.com/sparkwa/sparkwa/internal/spark"
	"github.com/sparkwa/sparkwa/internal/store"
	"github.com/sparkwa/sparkwa/internal/tui"
	"github.com/sparkwa/sparkwa/internal/whatsapp"
	"github.com/sparkwa/sparkwa/internal/worker"
	"golang.org/x/term"
)

var (
	Version   = "v1.0.0"
	GitCommit = "HEAD"
	BuildDate = "unknown"
)

func main() {
	var configPath string
	var cmd string

	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		if arg == "--config" && i+1 < len(os.Args) {
			configPath = os.Args[i+1]
			i++
		} else if strings.HasPrefix(arg, "--config=") {
			configPath = strings.TrimPrefix(arg, "--config=")
		} else if cmd == "" && !strings.HasPrefix(arg, "-") {
			cmd = arg
		} else if arg == "-h" || arg == "--help" || arg == "help" {
			cmd = "help"
		} else if arg == "-v" || arg == "--version" || arg == "version" {
			cmd = "version"
		}
	}

	if cmd == "" {
		cmd = "tui"
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	switch cmd {
	case "tui", "menu", "dashboard":
		if term.IsTerminal(int(os.Stdin.Fd())) && term.IsTerminal(int(os.Stdout.Fd())) {
			app := tui.NewApp(cfg)
			if err := app.Run(ctx); err != nil {
				log.Fatalf("TUI session encountered error: %v", err)
			}
			return
		}
		printUsage()

	case "wizard", "setup":
		if err := tui.RunSetupWizard(ctx, os.Stdin, os.Stdout, cfg); err != nil {
			log.Fatalf("Setup wizard error: %v", err)
		}

	case "version", "--version", "-v":
		fmt.Printf("SparkWA %s (%s) %s/%s built %s\n", Version, GitCommit, runtime.GOOS, runtime.GOARCH, BuildDate)

	case "init":
		runInit(cfg)

	case "login":
		runLogin(ctx, cfg)

	case "gmail-auth":
		runGmailAuth(ctx, cfg)

	case "doctor":
		runDoctor(ctx, cfg)

	case "status":
		runStatus(ctx, cfg)

	case "test-gmail":
		runTestGmail(ctx, cfg)

	case "run":
		runDaemon(ctx, cfg)

	case "help", "--help", "-h":
		printUsage()

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	tui.PrintLogo(os.Stdout, tui.GetTerminalWidth())
	fmt.Printf(`
Usage:
  sparkwa [command] [flags]

Interactive TUI & Wizards (Non-Technical Friendly):
  sparkwa                      Launch the interactive Black & White TUI Control Dashboard
  sparkwa wizard               Start the guided step-by-step setup wizard

Direct CLI Commands:
  run          Start the background bridge service (daemon mode)
  login        Pair WhatsApp Number 1 via terminal QR code
  gmail-auth   Authenticate with Google Gmail API (OAuth 2.0 helper)
  doctor       Run comprehensive health & capability diagnostics
  status       Inspect current bridge status and recent tasks
  test-gmail   Send a test verification message to the Spark Gmail inbox
  init         Initialize config and data directories with safe permissions
  version      Print application version and build details
  help         Display this command reference

Flags:
  --config     Explicit path to configuration file (default: /etc/sparkwa/config.env or ./config.env)

Documentation:
  https://github.com/sparkwa/sparkwa/tree/main/docs
`)
}

func runInit(cfg *config.Config) {
	fmt.Println("Initializing SparkWA environment...")

	// 1. Create data directory
	if err := os.MkdirAll(cfg.DataDir, 0700); err != nil {
		log.Fatalf("Failed to create data directory %s: %v", cfg.DataDir, err)
	}
	fmt.Printf("✓ Data directory ready: %s (mode 0700)\n", cfg.DataDir)

	// 2. Create sample config if none exists
	targetConfig := cfg.ConfigPath
	if targetConfig == "" {
		if os.Geteuid() == 0 {
			_ = os.MkdirAll("/etc/sparkwa", 0755)
			targetConfig = "/etc/sparkwa/config.env"
		} else {
			targetConfig = "./config.env"
		}
	}

	if _, err := os.Stat(targetConfig); os.IsNotExist(err) {
		template := `# SparkWA Configuration File
# WhatsApp Settings
BOT_PHONE_NUMBER="+1234567890"         # Number 1 (bot account)
AUTHORIZED_PHONE_NUMBER="+1987654321"  # Number 2 (your personal number, only permitted sender)

# Gmail Settings
BRIDGE_GMAIL_ADDRESS="bridge-bot@gmail.com"  # Mailbox used by the bridge
SPARK_GMAIL_ADDRESS="your-spark-account@gmail.com"  # Google account connected to Gemini Spark

# Runtime Paths & Polling
DATA_DIR="/var/lib/sparkwa"
GMAIL_POLL_INTERVAL="45s"
TASK_TIMEOUT="15m"
LOG_LEVEL="info"
`
		if err := os.WriteFile(targetConfig, []byte(template), 0600); err != nil {
			log.Fatalf("Failed to write config template to %s: %v", targetConfig, err)
		}
		fmt.Printf("✓ Created initial config file: %s (mode 0600)\n", targetConfig)
		fmt.Println("  Edit this file to specify your phone numbers and Gmail addresses.")
	} else {
		fmt.Printf("✓ Existing config found at: %s\n", targetConfig)
	}
}

func runLogin(ctx context.Context, cfg *config.Config) {
	fmt.Println("Initializing WhatsApp pairing client...")

	verifier, err := identity.NewVerifier(
		cfg.AuthorizedPhoneNumber,
		cfg.BotPhoneNumber,
		nil,
	)
	if err != nil {
		log.Fatalf("Identity initialization error: %v", err)
	}

	waClient, err := whatsapp.NewClient(ctx, cfg.WhatsAppDBPath, verifier, cfg.LogLevel)
	if err != nil {
		log.Fatalf("Failed to open WhatsApp session database: %v", err)
	}
	defer waClient.Close()

	if err := waClient.LoginQR(ctx); err != nil {
		log.Fatalf("WhatsApp login failed: %v", err)
	}

	fmt.Println("Waiting 3s for session registration to finalize on WhatsApp...")
	time.Sleep(3 * time.Second)

	fmt.Println("✓ WhatsApp pairing complete. Session saved in SQLite database.")
}

func runGmailAuth(ctx context.Context, cfg *config.Config) {
	if _, err := os.Stat(cfg.GmailCredentialsPath); os.IsNotExist(err) {
		fmt.Printf("Error: Google OAuth credentials file not found at:\n  %s\n\n", cfg.GmailCredentialsPath)
		fmt.Println("Please download your OAuth Client ID credentials JSON from the Google Cloud Console")
		fmt.Println("and save it to the path above. See docs/GMAIL_OAUTH_SETUP.md for full instructions.")
		os.Exit(1)
	}

	oauthCfg, err := gmail.LoadOAuth2Config(cfg.GmailCredentialsPath, cfg.OAuthLocalPort)
	if err != nil {
		log.Fatalf("Failed to load OAuth configuration: %v", err)
	}

	err = gmail.RunHeadlessOAuthHelper(ctx, oauthCfg, cfg.GmailTokenPath, cfg.OAuthLocalPort, os.Stdin, os.Stdout)
	if err != nil {
		log.Fatalf("OAuth authentication failed: %v", err)
	}

	// Test newly saved token
	gClient, err := gmail.NewClient(ctx, oauthCfg, cfg.GmailTokenPath)
	if err != nil {
		log.Fatalf("Failed to initialize Gmail client with new token: %v", err)
	}

	email, err := gClient.VerifyConnectivity(ctx, cfg.BridgeGmailAddress)
	if err != nil {
		fmt.Printf("⚠️ Warning: Connectivity check failed: %v\n", err)
	} else {
		fmt.Printf("✓ Successfully verified Gmail API connection for: %s\n", identity.MaskEmail(email))
	}
}

func runDoctor(ctx context.Context, cfg *config.Config) {
	s, err := store.Open(ctx, cfg.DBPath)
	if err != nil {
		log.Printf("Note: Could not open database (%v)", err)
	}
	defer func() {
		if s != nil {
			_ = s.Close()
		}
	}()

	var waClient *whatsapp.Client
	if verifier, err := identity.NewVerifier(cfg.AuthorizedPhoneNumber, cfg.BotPhoneNumber, nil); err == nil {
		waClient, _ = whatsapp.NewClient(ctx, cfg.WhatsAppDBPath, verifier, cfg.LogLevel)
		if waClient != nil {
			defer waClient.Close()
		}
	}

	var gClient *gmail.Client
	if oauthCfg, err := gmail.LoadOAuth2Config(cfg.GmailCredentialsPath, cfg.OAuthLocalPort); err == nil {
		gClient, _ = gmail.NewClient(ctx, oauthCfg, cfg.GmailTokenPath)
	}

	cm := spark.NewCapabilityManager(s, cfg)
	doc := health.NewDoctor(cfg, s, waClient, gClient, cm)

	results := doc.RunDiagnostics(ctx)
	allPass := health.PrintReport(results, os.Stdout)

	if !allPass {
		os.Exit(1)
	}
}

func runStatus(ctx context.Context, cfg *config.Config) {
	s, err := store.Open(ctx, cfg.DBPath)
	if err != nil {
		log.Fatalf("Cannot open database at %s: %v", cfg.DBPath, err)
	}
	defer s.Close()

	cm := spark.NewCapabilityManager(s, cfg)
	report, err := cm.Evaluate(ctx)
	if err != nil {
		log.Fatalf("Error evaluating capability: %v", err)
	}

	active, _ := s.GetActiveTasks(ctx)
	recent, _ := s.ListRecentTasks(ctx, 5)

	fmt.Println("================================================================================")
	fmt.Println("                           SPARKWA STATUS REPORT")
	fmt.Println("================================================================================")
	fmt.Printf("Capability State : %s\n", report.State)
	fmt.Printf("Description      : %s\n", report.Description)
	fmt.Printf("Can Submit       : %v | Can Receive: %v\n", report.CanSubmit, report.CanReceive)
	fmt.Printf("Next Steps       : %s\n", report.NextSteps)
	fmt.Println("--------------------------------------------------------------------------------")
	fmt.Printf("Active Queue     : %d task(s) currently awaiting Spark response\n", len(active))
	fmt.Printf("Recent Tasks     : %d recorded\n", len(recent))

	if len(recent) > 0 {
		fmt.Println("\nRecent Pipeline Events [Privacy-Safe]:")
		for _, t := range recent {
			displayID := t.ID
			if len(displayID) > 8 {
				displayID = displayID[:8]
			}
			fmt.Printf(" • [%-12s] Flow [%s] | Event %s (%s)\n", t.State, t.FlowTag(), displayID, t.CreatedAt.Format("15:04:05"))
		}
	}
	fmt.Println("================================================================================")
}

func runTestGmail(ctx context.Context, cfg *config.Config) {
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Config validation failed: %v", err)
	}

	oauthCfg, err := gmail.LoadOAuth2Config(cfg.GmailCredentialsPath, cfg.OAuthLocalPort)
	if err != nil {
		log.Fatalf("Failed to load OAuth config: %v", err)
	}

	gClient, err := gmail.NewClient(ctx, oauthCfg, cfg.GmailTokenPath)
	if err != nil {
		log.Fatalf("Failed to initialize Gmail client: %v", err)
	}

	email, err := gClient.VerifyConnectivity(ctx, cfg.BridgeGmailAddress)
	if err != nil {
		log.Fatalf("Gmail connectivity check failed: %v", err)
	}
	fmt.Printf("✓ Gmail API profile connected: %s\n", identity.MaskEmail(email))

	// Send a test read-only diagnostic ping to the Spark mailbox
	testID := fmt.Sprintf("test-%d", time.Now().Unix())
	subject := fmt.Sprintf("[SPARKWA:%s] SparkWA Diagnostics Ping", testID)
	body := fmt.Sprintf(
		"BEGIN SPARKWA COMMAND\n"+
			"Request ID: %s\n"+
			"Authorized sender: %s\n"+
			"Timestamp: %s\n"+
			"Command: READ_ONLY_DIAGNOSTIC_PING — Please respond with 'Diagnostics verified'.\n"+
			"END SPARKWA COMMAND\n",
		testID,
		identity.MaskPhone(cfg.AuthorizedPhoneNumber),
		time.Now().UTC().Format(time.RFC3339),
	)

	fmt.Printf("Sending test command to Spark inbox: %s ...\n", identity.MaskEmail(cfg.SparkGmailAddress))
	msgID, threadID, err := gClient.SendCommand(ctx, cfg.BridgeGmailAddress, cfg.SparkGmailAddress, subject, body, testID, cfg.AuthorizedPhoneNumber)
	if err != nil {
		log.Fatalf("Failed to send test command: %v", err)
	}

	fmt.Printf("✓ Command successfully sent!\n  Gmail Message ID: %s\n  Gmail Thread ID: %s\n", msgID, threadID)
	fmt.Println("Check your Spark-connected inbox to confirm the Gmail trigger fired.")
}

func runDaemon(ctx context.Context, cfg *config.Config) {
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Configuration validation failed:\n%v", err)
	}

	log.Printf("[Main] Starting SparkWA %s (%s)...", Version, GitCommit)

	// 1. Open SQLite database
	s, err := store.Open(ctx, cfg.DBPath)
	if err != nil {
		log.Fatalf("[Main] Failed to open database at %s: %v", cfg.DBPath, err)
	}
	defer s.Close()

	// 2. Initialize identity verifier
	verifier, err := identity.NewVerifier(cfg.AuthorizedPhoneNumber, cfg.BotPhoneNumber, nil)
	if err != nil {
		log.Fatalf("[Main] Identity verifier setup failed: %v", err)
	}

	// 3. Initialize WhatsApp client
	waClient, err := whatsapp.NewClient(ctx, cfg.WhatsAppDBPath, verifier, cfg.LogLevel)
	if err != nil {
		log.Fatalf("[Main] WhatsApp client setup failed: %v", err)
	}
	defer waClient.Close()

	// 4. Initialize Gmail client and poller
	oauthCfg, err := gmail.LoadOAuth2Config(cfg.GmailCredentialsPath, cfg.OAuthLocalPort)
	if err != nil {
		log.Fatalf("[Main] Failed to load Gmail OAuth credentials: %v", err)
	}

	gClient, err := gmail.NewClient(ctx, oauthCfg, cfg.GmailTokenPath)
	if err != nil {
		log.Fatalf("[Main] Failed to initialize Gmail client: %v", err)
	}

	gPoller := gmail.NewPoller(gClient, s, cfg.SparkGmailAddress, cfg.BridgeGmailAddress, cfg.SparkResponseMode)

	// 5. Capability Manager
	capManager := spark.NewCapabilityManager(s, cfg)

	// 6. Initialize and start worker engine
	eng := worker.NewEngine(cfg, s, gClient, gPoller, waClient, capManager)

	// Hook incoming WhatsApp messages to worker
	waClient.SetMessageHandler(func(ctx context.Context, msgID, senderPhone, text string) {
		eng.HandleInboundWhatsApp(ctx, msgID, senderPhone, text)
	})

	// Connect or pair if first time running
	if !waClient.IsLoggedIn() {
		log.Println("[Main] WhatsApp session not paired. Starting terminal QR code pairing...")
		if err := waClient.LoginQR(ctx); err != nil {
			log.Fatalf("[Main] WhatsApp pairing failed: %v", err)
		}
	} else {
		if err := waClient.Connect(); err != nil {
			log.Fatalf("[Main] WhatsApp connect failed: %v", err)
		}
	}
	defer waClient.Disconnect()

	if err := eng.Start(ctx); err != nil {
		log.Fatalf("[Main] Worker start failed: %v", err)
	}
	defer eng.Stop()

	log.Printf("[Main] SparkWA is running headlessly. Authorized user: %s", identity.MaskPhone(cfg.AuthorizedPhoneNumber))

	// Wait for shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	log.Printf("[Main] Received signal %s, initiating graceful shutdown...", sig)
}
