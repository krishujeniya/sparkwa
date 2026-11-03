package tui

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/sparkwa/sparkwa/internal/config"
	"github.com/sparkwa/sparkwa/internal/gmail"
	"github.com/sparkwa/sparkwa/internal/health"
	"github.com/sparkwa/sparkwa/internal/identity"
	"github.com/sparkwa/sparkwa/internal/message"
	"github.com/sparkwa/sparkwa/internal/spark"
	"github.com/sparkwa/sparkwa/internal/store"
	"github.com/sparkwa/sparkwa/internal/whatsapp"
	"github.com/sparkwa/sparkwa/internal/worker"
)

// App manages the interactive Terminal User Interface session.
type App struct {
	Config *config.Config
	In     io.Reader
	Out    io.Writer
}

// NewApp creates a new TUI App instance.
func NewApp(cfg *config.Config) *App {
	return &App{
		Config: cfg,
		In:     os.Stdin,
		Out:    os.Stdout,
	}
}

// Run starts the main interactive TUI loop.
func (a *App) Run(ctx context.Context) error {
	reader := bufio.NewReader(a.In)

	for {
		termWidth := GetTerminalWidth()
		ClearScreen(a.Out)

		// 1. Compute system status snapshot
		snap := a.evaluateSnapshot(ctx)

		// 2. Render Dashboard & Action Menu
		RenderDashboard(a.Out, snap, termWidth)

		// 3. Prompt user for selection
		PromptChoice(a.Out)
		input, err := reader.ReadString('\n')
		if err != nil {
			return nil
		}
		choice := strings.TrimSpace(input)

		switch choice {
		case "1", "wizard", "setup":
			ClearScreen(a.Out)
			_ = RunSetupWizard(ctx, a.In, a.Out, a.Config)

		case "2", "login", "pair", "qr":
			ClearScreen(a.Out)
			a.runWhatsAppPairing(ctx, reader)

		case "3", "gmail", "oauth", "auth":
			ClearScreen(a.Out)
			a.runGmailOAuth(ctx, reader)

		case "4", "doctor", "health", "diag":
			ClearScreen(a.Out)
			a.runDoctorDiagnostics(ctx, reader)

		case "5", "status", "queue", "jobs":
			ClearScreen(a.Out)
			a.runStatusAndQueue(ctx, reader)

		case "6", "test", "ping":
			ClearScreen(a.Out)
			a.runTestPing(ctx, reader)

		case "7", "run", "daemon", "start":
			ClearScreen(a.Out)
			a.runDaemonMode(ctx)

		case "8", "help", "doc", "docs":
			ClearScreen(a.Out)
			a.showDocumentation(reader)

		case "0", "exit", "quit", "q":
			ClearScreen(a.Out)
			fmt.Fprintln(a.Out, Bold("Thank you for using SparkWA. Goodbye!"))
			return nil

		default:
			// Invalid choice: loop repeats cleanly
		}
	}
}

func (a *App) evaluateSnapshot(ctx context.Context) StatusSnapshot {
	snap := StatusSnapshot{
		BotPhone:        a.Config.BotPhoneNumber,
		AuthPhone:       a.Config.AuthorizedPhoneNumber,
		BridgeEmail:     a.Config.BridgeGmailAddress,
		SparkEmail:      a.Config.SparkGmailAddress,
		ConfigPath:      a.Config.ConfigPath,
		CapabilityState: "NOT_CONFIGURED",
	}

	// Check WhatsApp database session
	if a.Config.WhatsAppDBPath != "" {
		if verifier, err := identity.NewVerifier(a.Config.AuthorizedPhoneNumber, a.Config.BotPhoneNumber, nil); err == nil {
			waClient, err := whatsapp.NewClient(ctx, a.Config.WhatsAppDBPath, verifier, "error")
			if err == nil && waClient != nil {
				snap.WhatsAppPaired = waClient.IsLoggedIn()
				_ = waClient.Close()
			}
		}
	}

	// Check Gmail OAuth token
	if _, err := os.Stat(a.Config.GmailTokenPath); err == nil {
		snap.GmailAuthed = true
	}

	// Check SQLite Task Store
	if s, err := store.Open(ctx, a.Config.DBPath); err == nil {
		cm := spark.NewCapabilityManager(s, a.Config)
		if rep, err := cm.Evaluate(ctx); err == nil {
			snap.CapabilityState = string(rep.State)
		}
		if active, err := s.GetActiveTasks(ctx); err == nil {
			snap.ActiveTasks = len(active)
		}
		_ = s.Close()
	}

	return snap
}

func (a *App) runWhatsAppPairing(ctx context.Context, reader *bufio.Reader) {
	fmt.Fprintln(a.Out, Bold("================================================================================"))
	fmt.Fprintln(a.Out, Bold("                     WHATSAPP TERMINAL QR PAIRING                               "))
	fmt.Fprintln(a.Out, Bold("================================================================================"))
	fmt.Fprintln(a.Out, "Instructions:")
	fmt.Fprintln(a.Out, "1. Open WhatsApp on Phone Number 1 (the bot account).")
	fmt.Fprintln(a.Out, "2. Go to Settings > Linked Devices > Link a Device.")
	fmt.Fprintln(a.Out, "3. Point your camera at the ASCII QR code that appears below.")
	fmt.Fprintln(a.Out, "================================================================================")
	fmt.Fprintln(a.Out)

	verifier, err := identity.NewVerifier(a.Config.AuthorizedPhoneNumber, a.Config.BotPhoneNumber, nil)
	if err != nil {
		fmt.Fprintf(a.Out, "%s Identity verifier error: %v\n", Inverse("[ERROR]"), err)
		a.waitEnter(reader)
		return
	}

	waClient, err := whatsapp.NewClient(ctx, a.Config.WhatsAppDBPath, verifier, a.Config.LogLevel)
	if err != nil {
		fmt.Fprintf(a.Out, "%s Failed to open session storage: %v\n", Inverse("[ERROR]"), err)
		a.waitEnter(reader)
		return
	}
	defer waClient.Close()

	if err := waClient.LoginQR(ctx); err != nil {
		fmt.Fprintf(a.Out, "%s WhatsApp pairing encountered error: %v\n", Inverse("[ERROR]"), err)
	} else {
		fmt.Fprintln(a.Out, "\nFinalizing session persistence...")
		time.Sleep(2 * time.Second)
		fmt.Fprintln(a.Out, Bold("✓ WhatsApp pairing complete! Session saved in SQLite database."))
	}

	a.waitEnter(reader)
}

func (a *App) runGmailOAuth(ctx context.Context, reader *bufio.Reader) {
	fmt.Fprintln(a.Out, Bold("================================================================================"))
	fmt.Fprintln(a.Out, Bold("                     GOOGLE GMAIL OAUTH 2.0 SETUP                               "))
	fmt.Fprintln(a.Out, Bold("================================================================================"))

	if _, err := os.Stat(a.Config.GmailCredentialsPath); os.IsNotExist(err) {
		fmt.Fprintf(a.Out, "%s Google OAuth credentials file not found at:\n  %s\n\n", Inverse("[FILE MISSING]"), a.Config.GmailCredentialsPath)
		fmt.Fprintln(a.Out, "Please download credentials.json from Google Cloud Console and place it at the path above.")
		fmt.Fprintln(a.Out, "See docs/GMAIL_OAUTH_SETUP.md for full step-by-step instructions.")
		a.waitEnter(reader)
		return
	}

	oauthCfg, err := gmail.LoadOAuth2Config(a.Config.GmailCredentialsPath, a.Config.OAuthLocalPort)
	if err != nil {
		fmt.Fprintf(a.Out, "%s Failed to parse credentials JSON: %v\n", Inverse("[ERROR]"), err)
		a.waitEnter(reader)
		return
	}

	err = gmail.RunHeadlessOAuthHelper(ctx, oauthCfg, a.Config.GmailTokenPath, a.Config.OAuthLocalPort, a.In, a.Out)
	if err != nil {
		fmt.Fprintf(a.Out, "%s OAuth authentication error: %v\n", Inverse("[ERROR]"), err)
		a.waitEnter(reader)
		return
	}

	// Verify newly generated token
	gClient, err := gmail.NewClient(ctx, oauthCfg, a.Config.GmailTokenPath)
	if err == nil {
		email, err := gClient.VerifyConnectivity(ctx, a.Config.BridgeGmailAddress)
		if err != nil {
			fmt.Fprintf(a.Out, "⚠️ Warning: Connectivity check failed: %v\n", err)
		} else {
			fmt.Fprintf(a.Out, "\n✓ Successfully verified Gmail API connection for: %s\n", Bold(email))
		}
	}

	a.waitEnter(reader)
}

func (a *App) runDoctorDiagnostics(ctx context.Context, reader *bufio.Reader) {
	s, err := store.Open(ctx, a.Config.DBPath)
	if err != nil {
		fmt.Fprintf(a.Out, "Notice: Database opening note (%v)\n", err)
	}
	defer func() {
		if s != nil {
			_ = s.Close()
		}
	}()

	var waClient *whatsapp.Client
	if verifier, err := identity.NewVerifier(a.Config.AuthorizedPhoneNumber, a.Config.BotPhoneNumber, nil); err == nil {
		waClient, _ = whatsapp.NewClient(ctx, a.Config.WhatsAppDBPath, verifier, "error")
		if waClient != nil {
			defer waClient.Close()
		}
	}

	var gClient *gmail.Client
	if oauthCfg, err := gmail.LoadOAuth2Config(a.Config.GmailCredentialsPath, a.Config.OAuthLocalPort); err == nil {
		gClient, _ = gmail.NewClient(ctx, oauthCfg, a.Config.GmailTokenPath)
	}

	cm := spark.NewCapabilityManager(s, a.Config)
	doc := health.NewDoctor(a.Config, s, waClient, gClient, cm)

	results := doc.RunDiagnostics(ctx)
	health.PrintReport(results, a.Out)

	a.waitEnter(reader)
}

func (a *App) runStatusAndQueue(ctx context.Context, reader *bufio.Reader) {
	s, err := store.Open(ctx, a.Config.DBPath)
	if err != nil {
		fmt.Fprintf(a.Out, "%s Cannot open database at %s: %v\n", Inverse("[ERROR]"), a.Config.DBPath, err)
		a.waitEnter(reader)
		return
	}
	defer s.Close()

	cm := spark.NewCapabilityManager(s, a.Config)
	report, err := cm.Evaluate(ctx)
	if err != nil {
		fmt.Fprintf(a.Out, "%s Error evaluating capability: %v\n", Inverse("[ERROR]"), err)
		a.waitEnter(reader)
		return
	}

	active, _ := s.GetActiveTasks(ctx)
	recent, _ := s.ListRecentTasks(ctx, 10)

	fmt.Fprintln(a.Out, Bold("================================================================================"))
	fmt.Fprintln(a.Out, Bold("                           SPARKWA STATUS REPORT                                "))
	fmt.Fprintln(a.Out, Bold("================================================================================"))
	fmt.Fprintf(a.Out, "Capability State : %s\n", Bold(string(report.State)))
	fmt.Fprintf(a.Out, "Description      : %s\n", report.Description)
	fmt.Fprintf(a.Out, "Can Submit       : %v | Can Receive: %v\n", report.CanSubmit, report.CanReceive)
	fmt.Fprintf(a.Out, "Next Steps       : %s\n", report.NextSteps)
	fmt.Fprintln(a.Out, "--------------------------------------------------------------------------------")
	fmt.Fprintf(a.Out, "Active Queue     : %d task(s) currently awaiting Spark response\n", len(active))
	fmt.Fprintf(a.Out, "Recent Tasks     : %d recorded\n", len(recent))

	if len(recent) > 0 {
		fmt.Fprintln(a.Out, "\nRecent Tasks:")
		for _, t := range recent {
			cmdExcerpt := strings.ReplaceAll(t.Command, "\n", " ")
			if len(cmdExcerpt) > 50 {
				cmdExcerpt = cmdExcerpt[:47] + "..."
			}
			fmt.Fprintf(a.Out, " • [%-12s] %s (%s) : %s\n", t.State, t.ID[:8], t.CreatedAt.Format("15:04:05"), cmdExcerpt)
		}
	}
	fmt.Fprintln(a.Out, Bold("================================================================================"))

	a.waitEnter(reader)
}

func (a *App) runTestPing(ctx context.Context, reader *bufio.Reader) {
	fmt.Fprintln(a.Out, Bold("================================================================================"))
	fmt.Fprintln(a.Out, Bold("                     OUTBOUND GMAIL DIAGNOSTICS TEST                            "))
	fmt.Fprintln(a.Out, Bold("================================================================================"))

	if err := a.Config.Validate(); err != nil {
		fmt.Fprintf(a.Out, "%s Configuration invalid: %v\n", Inverse("[CONFIG ERROR]"), err)
		a.waitEnter(reader)
		return
	}

	oauthCfg, err := gmail.LoadOAuth2Config(a.Config.GmailCredentialsPath, a.Config.OAuthLocalPort)
	if err != nil {
		fmt.Fprintf(a.Out, "%s Failed to load OAuth config: %v\n", Inverse("[ERROR]"), err)
		a.waitEnter(reader)
		return
	}

	gClient, err := gmail.NewClient(ctx, oauthCfg, a.Config.GmailTokenPath)
	if err != nil {
		fmt.Fprintf(a.Out, "%s Failed to initialize Gmail client: %v\n", Inverse("[ERROR]"), err)
		a.waitEnter(reader)
		return
	}

	email, err := gClient.VerifyConnectivity(ctx, a.Config.BridgeGmailAddress)
	if err != nil {
		fmt.Fprintf(a.Out, "%s Gmail connectivity error: %v\n", Inverse("[ERROR]"), err)
		a.waitEnter(reader)
		return
	}
	fmt.Fprintf(a.Out, "✓ Verified bridge Gmail profile: %s\n", Bold(email))

	testID := fmt.Sprintf("diag-%d", time.Now().Unix())
	subject := fmt.Sprintf("[SPARKWA:%s] Diagnostics Ping", testID)
	body := fmt.Sprintf(
		"BEGIN SPARKWA COMMAND\n"+
			"Request ID: %s\n"+
			"Authorized sender: %s\n"+
			"Timestamp: %s\n"+
			"Command: READ_ONLY_DIAGNOSTIC_PING — Please respond with 'Diagnostics verified'.\n"+
			"END SPARKWA COMMAND\n",
		testID,
		a.Config.AuthorizedPhoneNumber,
		time.Now().UTC().Format(time.RFC3339),
	)

	fmt.Fprintf(a.Out, "Dispatching test message to Spark mailbox: %s ...\n", Bold(a.Config.SparkGmailAddress))
	msgID, threadID, err := gClient.SendCommand(ctx, a.Config.BridgeGmailAddress, a.Config.SparkGmailAddress, subject, body, testID, a.Config.AuthorizedPhoneNumber)
	if err != nil {
		fmt.Fprintf(a.Out, "%s Failed to send test command: %v\n", Inverse("[ERROR]"), err)
	} else {
		fmt.Fprintln(a.Out, Bold("✓ Message successfully dispatched!"))
		fmt.Fprintf(a.Out, "  Message ID: %s\n  Thread ID:  %s\n", msgID, threadID)
		fmt.Fprintln(a.Out, "Check your Spark-connected inbox to confirm the Gmail trigger fired.")
	}

	a.waitEnter(reader)
}

func (a *App) runDaemonMode(ctx context.Context) {
	fmt.Fprintln(a.Out, Bold("================================================================================"))
	fmt.Fprintln(a.Out, Bold("                      STARTING SPARKWA BACKGROUND DAEMON                        "))
	fmt.Fprintln(a.Out, Bold("================================================================================"))
	fmt.Fprintln(a.Out, "The bridge is now running. All WhatsApp messages from your authorized number will")
	fmt.Fprintln(a.Out, "be processed and forwarded to Google Gemini Spark.")
	fmt.Fprintln(a.Out, Dim("To stop the daemon and return to the Control Menu, press Ctrl+C."))
	fmt.Fprintln(a.Out, Bold("================================================================================"))
	fmt.Fprintln(a.Out)

	if err := a.Config.Validate(); err != nil {
		fmt.Fprintf(a.Out, "%s Configuration error: %v\n", Inverse("[ERROR]"), err)
		time.Sleep(3 * time.Second)
		return
	}

	s, err := store.Open(ctx, a.Config.DBPath)
	if err != nil {
		fmt.Fprintf(a.Out, "%s Database error: %v\n", Inverse("[ERROR]"), err)
		time.Sleep(3 * time.Second)
		return
	}
	defer s.Close()

	verifier, err := identity.NewVerifier(a.Config.AuthorizedPhoneNumber, a.Config.BotPhoneNumber, nil)
	if err != nil {
		fmt.Fprintf(a.Out, "%s Verifier error: %v\n", Inverse("[ERROR]"), err)
		time.Sleep(3 * time.Second)
		return
	}

	waClient, err := whatsapp.NewClient(ctx, a.Config.WhatsAppDBPath, verifier, a.Config.LogLevel)
	if err != nil {
		fmt.Fprintf(a.Out, "%s WhatsApp client error: %v\n", Inverse("[ERROR]"), err)
		time.Sleep(3 * time.Second)
		return
	}
	defer waClient.Close()

	oauthCfg, err := gmail.LoadOAuth2Config(a.Config.GmailCredentialsPath, a.Config.OAuthLocalPort)
	if err != nil {
		fmt.Fprintf(a.Out, "%s Gmail OAuth config error: %v\n", Inverse("[ERROR]"), err)
		time.Sleep(3 * time.Second)
		return
	}

	gClient, err := gmail.NewClient(ctx, oauthCfg, a.Config.GmailTokenPath)
	if err != nil {
		fmt.Fprintf(a.Out, "%s Gmail client error: %v\n", Inverse("[ERROR]"), err)
		time.Sleep(3 * time.Second)
		return
	}

	gPoller := gmail.NewPoller(gClient, s, a.Config.SparkGmailAddress, a.Config.BridgeGmailAddress, a.Config.SparkResponseMode)
	cm := spark.NewCapabilityManager(s, a.Config)
	eng := worker.NewEngine(a.Config, s, gClient, gPoller, waClient, cm)

	waClient.SetMessageHandler(func(mCtx context.Context, msgID, senderPhone, text string) {
		eng.HandleInboundWhatsApp(mCtx, msgID, senderPhone, text)
	})

	if !waClient.IsLoggedIn() {
		fmt.Fprintf(a.Out, "%s WhatsApp is not paired yet. Please select [2] Pair WhatsApp first.\n", Inverse("[NOT PAIRED]"))
		time.Sleep(3 * time.Second)
		return
	}

	if err := waClient.Connect(); err != nil {
		fmt.Fprintf(a.Out, "%s WhatsApp connection error: %v\n", Inverse("[ERROR]"), err)
		time.Sleep(3 * time.Second)
		return
	}
	defer waClient.Disconnect()

	if err := eng.Start(ctx); err != nil {
		fmt.Fprintf(a.Out, "%s Worker start error: %v\n", Inverse("[ERROR]"), err)
		time.Sleep(3 * time.Second)
		return
	}
	defer eng.Stop()

	fmt.Fprintf(a.Out, "✓ %s Listening for authorized commands from: %s\n", Bold("SPARKWA ACTIVE:"), Bold(a.Config.AuthorizedPhoneNumber))

	// Catch Ctrl+C locally so we return to menu instead of terminating terminal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	signal.Stop(sigChan)

	fmt.Fprintln(a.Out, "\nGracefully stopped background service. Returning to Control Menu...")
	time.Sleep(1 * time.Second)
}

func (a *App) showDocumentation(reader *bufio.Reader) {
	fmt.Fprintln(a.Out, Bold("================================================================================"))
	fmt.Fprintln(a.Out, Bold("                     SPARKWA DOCUMENTATION & COMMAND GUIDE                      "))
	fmt.Fprintln(a.Out, Bold("================================================================================"))
	fmt.Fprintln(a.Out, message.FormatHelp())
	fmt.Fprintln(a.Out, "\nSystem Architecture:")
	fmt.Fprintln(a.Out, " • WhatsApp messages are validated with fail-closed sender authorization.")
	fmt.Fprintln(a.Out, " • Instructions are packaged in [SPARKWA:REQ_ID] envelopes and emailed to Spark.")
	fmt.Fprintln(a.Out, " • Spark processes tasks in Google Workspace and replies with results.")
	fmt.Fprintln(a.Out, " • Results are delivered directly back to your WhatsApp chat.")
	fmt.Fprintln(a.Out, Bold("================================================================================"))

	a.waitEnter(reader)
}

func (a *App) waitEnter(reader *bufio.Reader) {
	fmt.Fprint(a.Out, "\nPress Enter to return to the Control Menu...")
	_, _ = reader.ReadString('\n')
}
