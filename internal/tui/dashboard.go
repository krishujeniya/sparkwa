package tui

import (
	"fmt"
	"io"
)

// StatusSnapshot contains pre-evaluated status details for the TUI dashboard.
type StatusSnapshot struct {
	BotPhone        string
	AuthPhone       string
	BridgeEmail     string
	SparkEmail      string
	WhatsAppPaired  bool
	GmailAuthed     bool
	CapabilityState string
	ActiveTasks     int
	ConfigPath      string
}

// RenderDashboard prints the high-contrast monochromatic system dashboard and primary action menu.
func RenderDashboard(w io.Writer, snap StatusSnapshot, termWidth int) {
	// 1. Render Logo
	PrintLogo(w, termWidth)
	fmt.Fprintln(w)

	// 2. System Status Summary Card (Privacy-Safe Node Flow)
	var statusLines []string

	// Node 1: WhatsApp Bot
	waIndicator := IndicatorPending("NOT LINKED")
	waDesc := "Awaiting QR Pairing"
	if snap.WhatsAppPaired {
		waIndicator = IndicatorActive("ONLINE")
		waDesc = "Connected & Listening"
	}
	statusLines = append(statusLines, fmt.Sprintf("%-26s : %-26s %s", "Node [1] WhatsApp Bot", waDesc, waIndicator))

	// Node 2: Authorized User
	authIndicator := IndicatorActive("ALLOWED")
	authDesc := "Allowlisted Sender"
	if snap.AuthPhone == "" {
		authIndicator = IndicatorPending("CONFIG NEEDED")
		authDesc = "Unconfigured"
	}
	statusLines = append(statusLines, fmt.Sprintf("%-26s : %-26s %s", "Node [2] Authorized User", authDesc, authIndicator))

	// Node E: Gmail Bridge
	gIndicator := IndicatorPending("AUTH NEEDED")
	gDesc := "Token Required"
	if snap.GmailAuthed {
		gIndicator = IndicatorActive("READY")
		gDesc = "Authenticated (OAuth 2.0)"
	}
	statusLines = append(statusLines, fmt.Sprintf("%-26s : %-26s %s", "Node [E] Gmail Bridge", gDesc, gIndicator))

	// Node S: Spark Agent
	capStr := snap.CapabilityState
	if capStr == "" {
		capStr = "NOT_CONFIGURED"
	}
	sparkIndicator := IndicatorActive("ACTIVE")
	sparkDesc := "Integration Verified"
	if capStr == "NOT_CONFIGURED" {
		sparkIndicator = IndicatorPending("NOT READY")
		sparkDesc = "Setup Required"
	}
	statusLines = append(statusLines, fmt.Sprintf("%-26s : %-26s %s", "Node [S] Gemini Spark", sparkDesc, sparkIndicator))

	// Pipeline Data Flow
	statusLines = append(statusLines, "")
	statusLines = append(statusLines, fmt.Sprintf("%-26s : %s", "PIPELINE FLOW", Bold("[2 ➔ 1] ➔ [1 ➔ E] ➔ [E ➔ S] ➔ [S ➔ E] ➔ [E ➔ 2]")))

	// Queue State
	queueStr := fmt.Sprintf("%d active task(s) in queue", snap.ActiveTasks)
	statusLines = append(statusLines, fmt.Sprintf("%-26s : %-26s %s", "Persistent Queue", queueStr, IndicatorActive("HEALTHY")))

	boxWidth := termWidth
	if boxWidth > 82 {
		boxWidth = 82
	}
	fmt.Fprint(w, Box("SYSTEM INTEGRATION PIPELINE [PRIVACY-SAFE]", statusLines, boxWidth))
	fmt.Fprintln(w)

	// 3. Primary Actions Menu Card
	menuLines := []string{
		fmt.Sprintf("  %s  Walk through phone numbers, Gmail OAuth & WhatsApp QR", Bold("[1] Guided Setup Wizard")),
		fmt.Sprintf("  %s  Display terminal ASCII QR code to link Phone Number 1", Bold("[2] Pair WhatsApp")),
		fmt.Sprintf("  %s  Perform Google OAuth 2.0 loopback browser authorization", Bold("[3] Connect Gmail API")),
		fmt.Sprintf("  %s  Run 6-point diagnostics on config, permissions & DB", Bold("[4] System Diagnostics")),
		fmt.Sprintf("  %s  Inspect bridge capability report and recent tasks", Bold("[5] Queue & Status")),
		fmt.Sprintf("  %s  Dispatch safe read-only test ping to Spark mailbox", Bold("[6] Send Test Ping")),
		fmt.Sprintf("  %s  Launch production background listener service", Bold("[7] Start Bridge Daemon")),
		fmt.Sprintf("  %s  Show WhatsApp slash commands and operational guide", Bold("[8] Documentation & Help")),
		"",
		fmt.Sprintf("  %s  Exit to command shell", Dim("[0] Exit")),
	}

	fmt.Fprint(w, Box("CONTROL MENU — SELECT OPTION [0-8]", menuLines, boxWidth))
	fmt.Fprintln(w)
}

// PromptChoice prints an interactive command prompt asking the user for a selection.
func PromptChoice(w io.Writer) {
	fmt.Fprintf(w, "%s %s ", Inverse(" SPARKWA "), Bold("Enter choice [0-8] and press Enter:"))
}
