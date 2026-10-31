package health

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/sparkwa/sparkwa-lite/internal/config"
	"github.com/sparkwa/sparkwa-lite/internal/gmail"
	"github.com/sparkwa/sparkwa-lite/internal/spark"
	"github.com/sparkwa/sparkwa-lite/internal/store"
	"github.com/sparkwa/sparkwa-lite/internal/whatsapp"
)

// CheckStatus indicates whether a check passed, warned, or failed.
type CheckStatus string

const (
	StatusPass CheckStatus = "PASS"
	StatusWarn CheckStatus = "WARN"
	StatusFail CheckStatus = "FAIL"
)

// CheckResult represents the outcome of an individual diagnostic check.
type CheckResult struct {
	Category string      `json:"category"`
	Name     string      `json:"name"`
	Status   CheckStatus `json:"status"`
	Details  string      `json:"details"`
}

// Doctor aggregates diagnostics across all subsystems.
type Doctor struct {
	config     *config.Config
	store      *store.Store
	waClient   *whatsapp.Client
	gmailClient *gmail.Client
	capManager *spark.CapabilityManager
}

// NewDoctor creates a new diagnostic runner.
func NewDoctor(cfg *config.Config, s *store.Store, wa *whatsapp.Client, gc *gmail.Client, cm *spark.CapabilityManager) *Doctor {
	return &Doctor{
		config:      cfg,
		store:       s,
		waClient:    wa,
		gmailClient: gc,
		capManager:  cm,
	}
}

// RunDiagnostics executes all checks and returns the list of results.
func (d *Doctor) RunDiagnostics(ctx context.Context) []CheckResult {
	var results []CheckResult

	// 1. Config Check
	results = append(results, d.checkConfig()...)

	// 2. Filesystem Permissions Check
	results = append(results, d.checkPermissions()...)

	// 3. SQLite Database Health Check
	results = append(results, d.checkDatabase(ctx)...)

	// 4. WhatsApp Pairing Check
	results = append(results, d.checkWhatsApp()...)

	// 5. Gmail OAuth & API Connectivity Check
	results = append(results, d.checkGmail(ctx)...)

	// 6. Spark Capability Gate Check
	results = append(results, d.checkSpark(ctx)...)

	return results
}

// PrintReport writes a formatted diagnostic summary to the given writer.
func PrintReport(results []CheckResult, out io.Writer) bool {
	fmt.Fprintln(out, "================================================================================")
	fmt.Fprintln(out, "                     SPARKWA LITE SYSTEM DIAGNOSTICS")
	fmt.Fprintln(out, "================================================================================")

	allPass := true
	for _, r := range results {
		var symbol string
		switch r.Status {
		case StatusPass:
			symbol = "  [✓ PASS]"
		case StatusWarn:
			symbol = "  [! WARN]"
		case StatusFail:
			symbol = "  [✗ FAIL]"
			allPass = false
		}
		fmt.Fprintf(out, "%s %-16s %-24s : %s\n", symbol, r.Category, r.Name, r.Details)
	}

	fmt.Fprintln(out, "================================================================================")
	if allPass {
		fmt.Fprintln(out, "RESULT: All critical subsystems are healthy and ready for operation.")
	} else {
		fmt.Fprintln(out, "RESULT: One or more critical checks failed. Resolve issues before starting service.")
	}
	fmt.Fprintln(out, "================================================================================")

	return allPass
}

func (d *Doctor) checkConfig() []CheckResult {
	var res []CheckResult

	err := d.config.Validate()
	if err != nil {
		res = append(res, CheckResult{
			Category: "Configuration",
			Name:     "Validation",
			Status:   StatusFail,
			Details:  err.Error(),
		})
	} else {
		res = append(res, CheckResult{
			Category: "Configuration",
			Name:     "Validation",
			Status:   StatusPass,
			Details:  fmt.Sprintf("Bot: %s, User: %s, Bridge: %s, Spark: %s", d.config.BotPhoneNumber, d.config.AuthorizedPhoneNumber, d.config.BridgeGmailAddress, d.config.SparkGmailAddress),
		})
	}

	return res
}

func (d *Doctor) checkPermissions() []CheckResult {
	var res []CheckResult

	// Check DataDir
	dataDirInfo, err := os.Stat(d.config.DataDir)
	if err != nil {
		res = append(res, CheckResult{
			Category: "Permissions",
			Name:     "Data Directory",
			Status:   StatusWarn,
			Details:  fmt.Sprintf("Directory %s does not exist yet (will be created automatically)", d.config.DataDir),
		})
	} else {
		perm := dataDirInfo.Mode().Perm()
		if perm > 0755 {
			res = append(res, CheckResult{
				Category: "Permissions",
				Name:     "Data Directory",
				Status:   StatusWarn,
				Details:  fmt.Sprintf("Directory %s has overly permissive mode %o (recommended: 0700)", d.config.DataDir, perm),
			})
		} else {
			res = append(res, CheckResult{
				Category: "Permissions",
				Name:     "Data Directory",
				Status:   StatusPass,
				Details:  fmt.Sprintf("Directory %s exists with permissions %o", d.config.DataDir, perm),
			})
		}
	}

	// Check Token file permissions if exists
	if _, err := os.Stat(d.config.GmailTokenPath); err == nil {
		info, _ := os.Stat(d.config.GmailTokenPath)
		perm := info.Mode().Perm()
		if perm != 0600 {
			res = append(res, CheckResult{
				Category: "Permissions",
				Name:     "Gmail Token File",
				Status:   StatusWarn,
				Details:  fmt.Sprintf("%s has permissions %o (expected 0600)", d.config.GmailTokenPath, perm),
			})
		} else {
			res = append(res, CheckResult{
				Category: "Permissions",
				Name:     "Gmail Token File",
				Status:   StatusPass,
				Details:  fmt.Sprintf("%s has secure permissions 0600", d.config.GmailTokenPath),
			})
		}
	}

	return res
}

func (d *Doctor) checkDatabase(ctx context.Context) []CheckResult {
	var res []CheckResult

	if d.store == nil || d.store.DB() == nil {
		return append(res, CheckResult{
			Category: "Database",
			Name:     "SQLite Connection",
			Status:   StatusFail,
			Details:  "Database handle is not initialized",
		})
	}

	if err := d.store.DB().PingContext(ctx); err != nil {
		return append(res, CheckResult{
			Category: "Database",
			Name:     "SQLite Ping",
			Status:   StatusFail,
			Details:  fmt.Sprintf("Database ping failed: %v", err),
		})
	}

	var integrity string
	row := d.store.DB().QueryRowContext(ctx, "PRAGMA integrity_check;")
	if err := row.Scan(&integrity); err != nil || integrity != "ok" {
		res = append(res, CheckResult{
			Category: "Database",
			Name:     "Integrity Check",
			Status:   StatusFail,
			Details:  fmt.Sprintf("Integrity check failed: %v (%s)", err, integrity),
		})
	} else {
		res = append(res, CheckResult{
			Category: "Database",
			Name:     "Integrity Check",
			Status:   StatusPass,
			Details:  "SQLite database schema & index integrity is OK",
		})
	}

	return res
}

func (d *Doctor) checkWhatsApp() []CheckResult {
	var res []CheckResult

	if d.waClient == nil {
		return append(res, CheckResult{
			Category: "WhatsApp",
			Name:     "Client State",
			Status:   StatusFail,
			Details:  "WhatsApp client not initialized",
		})
	}

	if !d.waClient.IsLoggedIn() {
		res = append(res, CheckResult{
			Category: "WhatsApp",
			Name:     "Device Pairing",
			Status:   StatusFail,
			Details:  "Not paired to any WhatsApp account. Run 'sparkwa login' to scan terminal QR code.",
		})
	} else {
		res = append(res, CheckResult{
			Category: "WhatsApp",
			Name:     "Device Pairing",
			Status:   StatusPass,
			Details:  fmt.Sprintf("Paired to WhatsApp device %s", d.waClient.DeviceJID()),
		})
	}

	return res
}

func (d *Doctor) checkGmail(ctx context.Context) []CheckResult {
	var res []CheckResult

	// Check credentials JSON existence
	if _, err := os.Stat(d.config.GmailCredentialsPath); os.IsNotExist(err) {
		res = append(res, CheckResult{
			Category: "Gmail OAuth",
			Name:     "Credentials File",
			Status:   StatusFail,
			Details:  fmt.Sprintf("Missing credentials JSON at %s. Download from GCP Console.", d.config.GmailCredentialsPath),
		})
	} else {
		res = append(res, CheckResult{
			Category: "Gmail OAuth",
			Name:     "Credentials File",
			Status:   StatusPass,
			Details:  fmt.Sprintf("Found OAuth credentials at %s", filepath.Base(d.config.GmailCredentialsPath)),
		})
	}

	// Check token JSON existence
	if _, err := os.Stat(d.config.GmailTokenPath); os.IsNotExist(err) {
		res = append(res, CheckResult{
			Category: "Gmail OAuth",
			Name:     "Token File",
			Status:   StatusFail,
			Details:  fmt.Sprintf("No active token at %s. Run 'sparkwa gmail-auth'.", d.config.GmailTokenPath),
		})
		return res
	}

	if d.gmailClient == nil {
		res = append(res, CheckResult{
			Category: "Gmail API",
			Name:     "Connectivity",
			Status:   StatusWarn,
			Details:  "Gmail API client not yet initialized",
		})
		return res
	}

	// Verify live API connectivity
	email, err := d.gmailClient.VerifyConnectivity(ctx, d.config.BridgeGmailAddress)
	if err != nil {
		res = append(res, CheckResult{
			Category: "Gmail API",
			Name:     "Connectivity",
			Status:   StatusFail,
			Details:  fmt.Sprintf("Gmail API profile check failed: %v", err),
		})
	} else {
		res = append(res, CheckResult{
			Category: "Gmail API",
			Name:     "Connectivity",
			Status:   StatusPass,
			Details:  fmt.Sprintf("Successfully connected as %s", email),
		})
	}

	return res
}

func (d *Doctor) checkSpark(ctx context.Context) []CheckResult {
	var res []CheckResult

	if d.capManager == nil {
		return append(res, CheckResult{
			Category: "Gemini Spark",
			Name:     "Integration Gate",
			Status:   StatusFail,
			Details:  "Capability manager not initialized",
		})
	}

	report, err := d.capManager.Evaluate(ctx)
	if err != nil {
		return append(res, CheckResult{
			Category: "Gemini Spark",
			Name:     "Integration Gate",
			Status:   StatusFail,
			Details:  fmt.Sprintf("Failed to evaluate capability state: %v", err),
		})
	}

	status := StatusPass
	if report.State == config.CapNotConfigured {
		status = StatusFail
	} else if report.State == config.CapSubmissionOnly || report.State == config.CapGmailConnected || report.State == config.CapSparkTriggerConfigured {
		status = StatusWarn
	}

	res = append(res, CheckResult{
		Category: "Gemini Spark",
		Name:     "Capability State",
		Status:   status,
		Details:  fmt.Sprintf("[%s] %s %s", report.State, report.Description, report.NextSteps),
	})

	return res
}
