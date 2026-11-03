package spark

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/sparkwa/sparkwa/internal/config"
	"github.com/sparkwa/sparkwa/internal/store"
)

func TestCapabilityManager(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_cap.db")

	s, err := store.Open(context.Background(), dbPath)
	if err != nil {
		t.Fatalf("failed to open store: %v", err)
	}
	defer s.Close()

	cfg := config.DefaultConfig()
	cm := NewCapabilityManager(s, cfg)
	ctx := context.Background()

	// 1. Initial state should be NOT_CONFIGURED
	st, err := cm.GetCurrentState(ctx)
	if err != nil {
		t.Fatalf("GetCurrentState error: %v", err)
	}
	if st != config.CapNotConfigured {
		t.Errorf("expected initial state %s, got %s", config.CapNotConfigured, st)
	}

	report, err := cm.Evaluate(ctx)
	if err != nil {
		t.Fatalf("Evaluate error: %v", err)
	}
	if report.CanSubmit || report.CanReceive {
		t.Errorf("NOT_CONFIGURED should not permit submit or receive")
	}

	// 2. Transition to GMAIL_CONNECTED
	if err := cm.SetState(ctx, config.CapGmailConnected); err != nil {
		t.Fatalf("SetState error: %v", err)
	}
	report, _ = cm.Evaluate(ctx)
	if !report.CanSubmit || report.CanReceive {
		t.Errorf("GMAIL_CONNECTED should allow submit but not receive")
	}

	// 3. Transition to SUBMISSION_ONLY
	if err := cm.SetState(ctx, config.CapSubmissionOnly); err != nil {
		t.Fatalf("SetState error: %v", err)
	}
	report, _ = cm.Evaluate(ctx)
	if !report.CanSubmit || report.CanReceive {
		t.Errorf("SUBMISSION_ONLY should allow submit only")
	}

	// 4. Transition to FULLY_VERIFIED
	if err := cm.SetState(ctx, config.CapFullyVerified); err != nil {
		t.Fatalf("SetState error: %v", err)
	}
	report, _ = cm.Evaluate(ctx)
	if !report.CanSubmit || !report.CanReceive {
		t.Errorf("FULLY_VERIFIED should allow both submit and receive")
	}
}
