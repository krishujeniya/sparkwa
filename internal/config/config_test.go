package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{
			name: "Valid configuration",
			cfg: Config{
				BotPhoneNumber:        "+1234567890",
				AuthorizedPhoneNumber: "+1987654321",
				BridgeGmailAddress:    "bridge@example.com",
				SparkGmailAddress:     "spark@example.com",
				GmailPollInterval:    45 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "Missing bot phone",
			cfg: Config{
				AuthorizedPhoneNumber: "+1987654321",
				BridgeGmailAddress:    "bridge@example.com",
				SparkGmailAddress:     "spark@example.com",
				GmailPollInterval:    45 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "Invalid E.164 without plus",
			cfg: Config{
				BotPhoneNumber:        "1234567890",
				AuthorizedPhoneNumber: "+1987654321",
				BridgeGmailAddress:    "bridge@example.com",
				SparkGmailAddress:     "spark@example.com",
				GmailPollInterval:    45 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "Invalid bridge email",
			cfg: Config{
				BotPhoneNumber:        "+1234567890",
				AuthorizedPhoneNumber: "+1987654321",
				BridgeGmailAddress:    "invalid-email",
				SparkGmailAddress:     "spark@example.com",
				GmailPollInterval:    45 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "Poll interval too short",
			cfg: Config{
				BotPhoneNumber:        "+1234567890",
				AuthorizedPhoneNumber: "+1987654321",
				BridgeGmailAddress:    "bridge@example.com",
				SparkGmailAddress:     "spark@example.com",
				GmailPollInterval:    2 * time.Second,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfigLoadFile(t *testing.T) {
	tempDir := t.TempDir()
	envPath := filepath.Join(tempDir, "test.env")

	content := `
BOT_PHONE_NUMBER="+1112223333"
AUTHORIZED_PHONE_NUMBER="+14445556666"
BRIDGE_GMAIL_ADDRESS="testbridge@gmail.com"
SPARK_GMAIL_ADDRESS="testspark@gmail.com"
GMAIL_POLL_INTERVAL="30s"
LOG_LEVEL="debug"
`
	if err := os.WriteFile(envPath, []byte(content), 0600); err != nil {
		t.Fatalf("failed to write test env: %v", err)
	}

	// Clean env before test
	os.Unsetenv("BOT_PHONE_NUMBER")
	os.Unsetenv("AUTHORIZED_PHONE_NUMBER")
	os.Unsetenv("BRIDGE_GMAIL_ADDRESS")
	os.Unsetenv("SPARK_GMAIL_ADDRESS")
	os.Unsetenv("GMAIL_POLL_INTERVAL")
	os.Unsetenv("LOG_LEVEL")

	cfg, err := Load(envPath)
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}

	if cfg.BotPhoneNumber != "+1112223333" {
		t.Errorf("expected bot phone +1112223333, got %s", cfg.BotPhoneNumber)
	}
	if cfg.AuthorizedPhoneNumber != "+14445556666" {
		t.Errorf("expected auth phone +14445556666, got %s", cfg.AuthorizedPhoneNumber)
	}
	if cfg.GmailPollInterval != 30*time.Second {
		t.Errorf("expected 30s poll interval, got %v", cfg.GmailPollInterval)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("expected log level debug, got %s", cfg.LogLevel)
	}

	if err := cfg.Validate(); err != nil {
		t.Errorf("Validate() unexpected error: %v", err)
	}
}
