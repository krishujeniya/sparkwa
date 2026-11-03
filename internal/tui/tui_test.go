package tui

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/sparkwa/sparkwa/internal/config"
)

func TestMonochromeThemeTokens(t *testing.T) {
	b := Bold("test")
	if !strings.HasPrefix(b, "\033[1m") || !strings.HasSuffix(b, "\033[0m") {
		t.Errorf("unexpected Bold format: %q", b)
	}

	d := Dim("subtext")
	if !strings.HasPrefix(d, "\033[2m") || !strings.HasSuffix(d, "\033[0m") {
		t.Errorf("unexpected Dim format: %q", d)
	}

	inv := Inverse("inverted")
	if !strings.HasPrefix(inv, "\033[7m") || !strings.HasSuffix(inv, "\033[0m") {
		t.Errorf("unexpected Inverse format: %q", inv)
	}

	badge := Badge("OK")
	if !strings.Contains(badge, " OK ") {
		t.Errorf("unexpected Badge format: %q", badge)
	}

	active := IndicatorActive("ACTIVE")
	if !strings.Contains(active, "● ACTIVE") {
		t.Errorf("unexpected IndicatorActive format: %q", active)
	}

	pending := IndicatorPending("PENDING")
	if !strings.Contains(pending, "○ PENDING") {
		t.Errorf("unexpected IndicatorPending format: %q", pending)
	}

	warn := IndicatorWarn("ALERT")
	if !strings.Contains(warn, "! ALERT") {
		t.Errorf("unexpected IndicatorWarn format: %q", warn)
	}
}

func TestVisibleLength(t *testing.T) {
	plain := "Hello World"
	if visibleLength(plain) != 11 {
		t.Errorf("expected 11, got %d", visibleLength(plain))
	}

	styled := Bold("Hello") + " " + Dim("World")
	if visibleLength(styled) != 11 {
		t.Errorf("expected 11 for styled text, got %d", visibleLength(styled))
	}
}

func TestBoxFraming(t *testing.T) {
	lines := []string{"Line 1", "Line 2 is longer"}
	box := Box("STATUS", lines, 40)

	if !strings.Contains(box, "┌─ STATUS") {
		t.Errorf("box missing top border with title: %s", box)
	}
	if !strings.Contains(box, "│ Line 1") {
		t.Errorf("box missing content: %s", box)
	}
	if !strings.Contains(box, "└") {
		t.Errorf("box missing bottom border: %s", box)
	}
}

func TestLogoRendering(t *testing.T) {
	// Full width test
	var bufFull bytes.Buffer
	PrintLogo(&bufFull, 82)
	fullOut := bufFull.String()
	if !strings.Contains(fullOut, "THE HEADLESS WHATSAPP TO GEMINI SPARK BRIDGE") {
		t.Errorf("expected full banner in 82-column output, got: %s", fullOut)
	}
	if !strings.Contains(fullOut, "By Krish Ujeniya") {
		t.Errorf("expected author credit in full banner, got: %s", fullOut)
	}

	// Compact width test (< 80 cols)
	var bufCompact bytes.Buffer
	PrintLogo(&bufCompact, 60)
	compactOut := bufCompact.String()
	if !strings.Contains(compactOut, "SPARKWA") {
		t.Errorf("expected compact banner in 60-column output, got: %s", compactOut)
	}
	if !strings.Contains(compactOut, "By Krish Ujeniya") {
		t.Errorf("expected author credit in compact banner, got: %s", compactOut)
	}
}

func TestRenderDashboard(t *testing.T) {
	snap := StatusSnapshot{
		BotPhone:        "+1234567890",
		AuthPhone:       "+1987654321",
		BridgeEmail:     "bridge@example.com",
		SparkEmail:      "spark@example.com",
		WhatsAppPaired:  true,
		GmailAuthed:     true,
		CapabilityState: "FULLY_VERIFIED",
		ActiveTasks:     2,
	}

	var buf bytes.Buffer
	RenderDashboard(&buf, snap, 82)
	out := buf.String()

	if !strings.Contains(out, "SYSTEM INTEGRATION PIPELINE [PRIVACY-SAFE]") {
		t.Errorf("missing snapshot title in dashboard: %s", out)
	}
	if !strings.Contains(out, "Node [1] WhatsApp Bot") {
		t.Errorf("missing node 1 in dashboard: %s", out)
	}
	if !strings.Contains(out, "Node [E] Gmail Bridge") {
		t.Errorf("missing node E in dashboard: %s", out)
	}
	if !strings.Contains(out, "[● ONLINE]") {
		t.Errorf("missing online indicator in dashboard: %s", out)
	}
	if !strings.Contains(out, "[1] Guided Setup Wizard") {
		t.Errorf("missing menu option 1 in dashboard: %s", out)
	}
}

func TestSetupWizard(t *testing.T) {
	tempDir := t.TempDir()
	cfgPath := filepath.Join(tempDir, "wizard_test.env")

	cfg := &config.Config{
		DataDir:              tempDir,
		GmailCredentialsPath: filepath.Join(tempDir, "credentials.json"),
		GmailTokenPath:       filepath.Join(tempDir, "token.json"),
		GmailPollInterval:    45 * time.Second,
		TaskTimeout:          15 * time.Minute,
		LogLevel:             "info",
		ConfigPath:           cfgPath,
	}

	// Simulated user input:
	// 1. Bot Phone: +1234567890
	// 2. Personal Phone: +1987654321
	// 3. Bridge Email: bridge@example.com
	// 4. Spark Email: spark@example.com
	// 5. Confirm Save: y
	// 6. Return Enter: \n
	userInput := "+1234567890\n+1987654321\nbridge@example.com\nspark@example.com\ny\n\n"
	in := strings.NewReader(userInput)
	var out bytes.Buffer

	err := RunSetupWizard(context.Background(), in, &out, cfg)
	if err != nil {
		t.Fatalf("wizard returned error: %v", err)
	}

	// Verify updated config
	if cfg.BotPhoneNumber != "+1234567890" {
		t.Errorf("expected bot phone +1234567890, got %s", cfg.BotPhoneNumber)
	}
	if cfg.AuthorizedPhoneNumber != "+1987654321" {
		t.Errorf("expected auth phone +1987654321, got %s", cfg.AuthorizedPhoneNumber)
	}

	// Verify saved config file on disk
	content, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatalf("failed to read saved config file: %v", err)
	}
	if !strings.Contains(string(content), "BOT_PHONE_NUMBER=\"+1234567890\"") {
		t.Errorf("saved config does not match: %s", string(content))
	}
}
