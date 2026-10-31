package health

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/sparkwa/sparkwa-lite/internal/config"
	"github.com/sparkwa/sparkwa-lite/internal/spark"
	"github.com/sparkwa/sparkwa-lite/internal/store"
)

func TestDoctorDiagnostics(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "doctor_test.db")

	s, err := store.Open(context.Background(), dbPath)
	if err != nil {
		t.Fatalf("failed to open store: %v", err)
	}
	defer s.Close()

	cfg := &config.Config{
		BotPhoneNumber:        "+12345678900",
		AuthorizedPhoneNumber: "+19876543210",
		BridgeGmailAddress:    "bridge@example.com",
		SparkGmailAddress:     "spark@example.com",
		DataDir:               tempDir,
		GmailCredentialsPath:  filepath.Join(tempDir, "credentials.json"),
		GmailTokenPath:        filepath.Join(tempDir, "token.json"),
		DBPath:                dbPath,
		GmailPollInterval:    45 * time.Second,
	}

	cm := spark.NewCapabilityManager(s, cfg)
	doc := NewDoctor(cfg, s, nil, nil, cm)

	ctx := context.Background()
	results := doc.RunDiagnostics(ctx)

	if len(results) == 0 {
		t.Fatalf("expected diagnostic results, got 0")
	}

	var buf bytes.Buffer
	allPass := PrintReport(results, &buf)

	output := buf.String()
	if !strings.Contains(output, "SPARKWA LITE SYSTEM DIAGNOSTICS") {
		t.Errorf("missing title in report: %s", output)
	}

	// Because whatsapp and gmail tokens are mock/missing in unit test, allPass should be false
	if allPass {
		t.Errorf("expected allPass to be false for unconfigured test environment")
	}
}
