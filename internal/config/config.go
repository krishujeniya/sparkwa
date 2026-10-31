package config

import (
	"bufio"
	"errors"
	"fmt"
	"net/mail"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// CapabilityState represents the verified capability of the Gemini Spark bridge.
type CapabilityState string

const (
	CapNotConfigured           CapabilityState = "NOT_CONFIGURED"
	CapGmailConnected          CapabilityState = "GMAIL_CONNECTED"
	CapSparkTriggerConfigured  CapabilityState = "SPARK_TRIGGER_CONFIGURED"
	CapResponsePathVerified    CapabilityState = "RESPONSE_PATH_VERIFIED"
	CapSubmissionOnly          CapabilityState = "SUBMISSION_ONLY"
	CapFullyVerified           CapabilityState = "FULLY_VERIFIED"
)

// Config holds all runtime configuration for SparkWA Lite.
type Config struct {
	// WhatsApp settings
	BotPhoneNumber        string // E.164, Number 1 (bot account)
	AuthorizedPhoneNumber string // E.164, Number 2 (authorized user)

	// Gmail settings
	BridgeGmailAddress string // Mailbox used by SparkWA to send commands & receive replies
	SparkGmailAddress  string // Google account where Gemini Spark task runs

	// Paths
	DataDir              string // Directory for databases & runtime state (default: /var/lib/sparkwa-lite)
	ConfigPath           string // Path to config file loaded
	GmailCredentialsPath string // Path to GCP OAuth client credentials JSON
	GmailTokenPath       string // Path to stored OAuth refresh/access token JSON
	DBPath               string // Path to application SQLite database
	WhatsAppDBPath       string // Path to whatsmeow session SQLite database

	// Runtime settings
	GmailPollInterval time.Duration // Interval to poll Gmail for replies (default: 45s)
	TaskTimeout       time.Duration // Maximum time before a task expires (default: 15m)
	MaxTaskRetries    int           // Retries before failing a task (default: 3)
	MaxConcurrentJobs int           // Concurrently active jobs (default: 5)
	MaxMessageLength  int           // Maximum inbound/outbound message length (default: 4000)
	LogLevel          string        // debug, info, warn, error (default: info)
	OAuthLocalPort    int           // Loopback HTTP port for OAuth callback (default: 8085)
	SparkResponseMode string        // draft_or_email, email_only, draft_only (default: draft_or_email)
}

// DefaultConfig returns safe defaults for production VPS deployment.
func DefaultConfig() *Config {
	dataDir := "/var/lib/sparkwa-lite"
	// Fall back to local data directory if not running with root/service privileges
	if os.Geteuid() != 0 {
		home, err := os.UserHomeDir()
		if err == nil {
			localDir := filepath.Join(home, ".local", "share", "sparkwa-lite")
			dataDir = localDir
		} else {
			dataDir = "./data"
		}
	}

	return &Config{
		DataDir:              dataDir,
		GmailCredentialsPath: filepath.Join(dataDir, "credentials.json"),
		GmailTokenPath:       filepath.Join(dataDir, "token.json"),
		DBPath:               filepath.Join(dataDir, "sparkwa.db"),
		WhatsAppDBPath:       filepath.Join(dataDir, "whatsapp.db"),
		GmailPollInterval:    45 * time.Second,
		TaskTimeout:          15 * time.Minute,
		MaxTaskRetries:       3,
		MaxConcurrentJobs:    5,
		MaxMessageLength:     4000,
		LogLevel:             "info",
		OAuthLocalPort:       8085,
		SparkResponseMode:    "draft_or_email",
	}
}

// Load loads configuration from environment variables and an optional file.
// Priority: Environment variables > Specified file > Default search paths.
func Load(explicitFile string) (*Config, error) {
	cfg := DefaultConfig()

	// Try loading from file if found
	var foundPath string
	candidates := []string{}
	if explicitFile != "" {
		candidates = append(candidates, explicitFile)
	}
	candidates = append(candidates,
		"/etc/sparkwa-lite/config.env",
		"./config.env",
		".env",
	)

	for _, p := range candidates {
		if p == "" {
			continue
		}
		if _, err := os.Stat(p); err == nil {
			foundPath = p
			break
		}
	}

	if foundPath != "" {
		cfg.ConfigPath = foundPath
		if err := loadEnvFile(foundPath); err != nil {
			return nil, fmt.Errorf("failed to parse config file %s: %w", foundPath, err)
		}
	}

	// Apply environment variables
	if v := os.Getenv("BOT_PHONE_NUMBER"); v != "" {
		cfg.BotPhoneNumber = strings.TrimSpace(v)
	}
	if v := os.Getenv("AUTHORIZED_PHONE_NUMBER"); v != "" {
		cfg.AuthorizedPhoneNumber = strings.TrimSpace(v)
	}
	if v := os.Getenv("BRIDGE_GMAIL_ADDRESS"); v != "" {
		cfg.BridgeGmailAddress = strings.TrimSpace(v)
	}
	if v := os.Getenv("SPARK_GMAIL_ADDRESS"); v != "" {
		cfg.SparkGmailAddress = strings.TrimSpace(v)
	}
	if v := os.Getenv("DATA_DIR"); v != "" {
		cfg.DataDir = strings.TrimSpace(v)
		// Update dependent paths if not explicitly overridden
		cfg.GmailCredentialsPath = filepath.Join(cfg.DataDir, "credentials.json")
		cfg.GmailTokenPath = filepath.Join(cfg.DataDir, "token.json")
		cfg.DBPath = filepath.Join(cfg.DataDir, "sparkwa.db")
		cfg.WhatsAppDBPath = filepath.Join(cfg.DataDir, "whatsapp.db")
	}
	if v := os.Getenv("GMAIL_CREDENTIALS_FILE"); v != "" {
		cfg.GmailCredentialsPath = strings.TrimSpace(v)
	}
	if v := os.Getenv("GMAIL_TOKEN_FILE"); v != "" {
		cfg.GmailTokenPath = strings.TrimSpace(v)
	}
	if v := os.Getenv("DB_PATH"); v != "" {
		cfg.DBPath = strings.TrimSpace(v)
	}
	if v := os.Getenv("WHATSAPP_DB_PATH"); v != "" {
		cfg.WhatsAppDBPath = strings.TrimSpace(v)
	}
	if v := os.Getenv("GMAIL_POLL_INTERVAL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.GmailPollInterval = d
		}
	}
	if v := os.Getenv("TASK_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.TaskTimeout = d
		}
	}
	if v := os.Getenv("LOG_LEVEL"); v != "" {
		cfg.LogLevel = strings.ToLower(strings.TrimSpace(v))
	}
	if v := os.Getenv("SPARK_RESPONSE_MODE"); v != "" {
		cfg.SparkResponseMode = strings.TrimSpace(v)
	}

	return cfg, nil
}

// Validate checks whether the required configuration fields are valid for daemon execution.
func (c *Config) Validate() error {
	var errs []string

	if c.BotPhoneNumber == "" {
		errs = append(errs, "BOT_PHONE_NUMBER is required (e.g. +1234567890)")
	} else if !strings.HasPrefix(c.BotPhoneNumber, "+") || len(c.BotPhoneNumber) < 8 {
		errs = append(errs, "BOT_PHONE_NUMBER must be in E.164 format (+<country><digits>)")
	}

	if c.AuthorizedPhoneNumber == "" {
		errs = append(errs, "AUTHORIZED_PHONE_NUMBER is required (e.g. +1987654321)")
	} else if !strings.HasPrefix(c.AuthorizedPhoneNumber, "+") || len(c.AuthorizedPhoneNumber) < 8 {
		errs = append(errs, "AUTHORIZED_PHONE_NUMBER must be in E.164 format (+<country><digits>)")
	}

	if c.BridgeGmailAddress == "" {
		errs = append(errs, "BRIDGE_GMAIL_ADDRESS is required")
	} else if _, err := mail.ParseAddress(c.BridgeGmailAddress); err != nil {
		errs = append(errs, fmt.Sprintf("invalid BRIDGE_GMAIL_ADDRESS: %v", err))
	}

	if c.SparkGmailAddress == "" {
		errs = append(errs, "SPARK_GMAIL_ADDRESS is required")
	} else if _, err := mail.ParseAddress(c.SparkGmailAddress); err != nil {
		errs = append(errs, fmt.Sprintf("invalid SPARK_GMAIL_ADDRESS: %v", err))
	}

	if c.GmailPollInterval < 10*time.Second {
		errs = append(errs, "GMAIL_POLL_INTERVAL must be at least 10 seconds to avoid API quota drain")
	}

	if len(errs) > 0 {
		return errors.New("configuration validation failed:\n - " + strings.Join(errs, "\n - "))
	}

	return nil
}

// loadEnvFile reads a simple KEY=VALUE file and sets environment variables if not already set.
func loadEnvFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		// Strip surrounding quotes
		if (strings.HasPrefix(val, "\"") && strings.HasSuffix(val, "\"")) ||
			(strings.HasPrefix(val, "'") && strings.HasSuffix(val, "'")) {
			val = val[1 : len(val)-1]
		}
		// Don't overwrite existing environment variables
		if os.Getenv(key) == "" {
			_ = os.Setenv(key, val)
		}
	}
	return scanner.Err()
}
