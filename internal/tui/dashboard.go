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

	// 2. System Status Summary Card
	var statusLines []string

	// WhatsApp Status
	waIndicator := IndicatorPending("NOT LINKED")
	if snap.WhatsAppPaired {
		waIndicator = IndicatorActive("CONNECTED")
	}
	botPhoneStr := snap.BotPhone
	if botPhoneStr == "" {
		botPhoneStr = "Unconfigured"
	}
	statusLines = append(statusLines, fmt.Sprintf("%-26s : %-26s %s", "WhatsApp Bot (Number 1)", botPhoneStr, waIndicator))

	// Authorized User
	authPhoneStr := snap.AuthPhone
	if authPhoneStr == "" {
		authPhoneStr = "Unconfigured"
	}
	statusLines = append(statusLines, fmt.Sprintf("%-26s : %-26s %s", "Authorized User (Number 2)", authPhoneStr, IndicatorActive("ALLOWLISTED")))

	// Gmail Bridge Status
	gIndicator := IndicatorPending("TOKEN NEEDED")
	if snap.GmailAuthed {
		gIndicator = IndicatorActive("CONNECTED")
	}
	bridgeEmailStr := snap.BridgeEmail
	if bridgeEmailStr == "" {
		bridgeEmailStr = "Unconfigured"
	}
	if len(bridgeEmailStr) > 26 {
		bridgeEmailStr = bridgeEmailStr[:23] + "..."
	}
	statusLines = append(statusLines, fmt.Sprintf("%-26s : %-26s %s", "Gmail Bridge Mailbox", bridgeEmailStr, gIndicator))

	// Spark Mailbox
	sparkEmailStr := snap.SparkEmail
	if sparkEmailStr == "" {
		sparkEmailStr = "Unconfigured"
	}
	if len(sparkEmailStr) > 26 {
		sparkEmailStr = sparkEmailStr[:23] + "..."
	}
	capStr := snap.CapabilityState
	if capStr == "" {
		capStr = "NOT_CONFIGURED"
	}
	statusLines = append(statusLines, fmt.Sprintf("%-26s : %-26s %s", "Spark Agent Destination", sparkEmailStr, Bold("["+capStr+"]")))

	// Queue State
	queueStr := fmt.Sprintf("%d task(s) active in queue", snap.ActiveTasks)
	statusLines = append(statusLines, fmt.Sprintf("%-26s : %-26s %s", "SQLite Persistent Queue", queueStr, IndicatorActive("HEALTHY")))

	boxWidth := termWidth
	if boxWidth > 82 {
		boxWidth = 82
	}
	fmt.Fprint(w, Box("SYSTEM HEALTH & INTEGRATION SNAPSHOT", statusLines, boxWidth))
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
