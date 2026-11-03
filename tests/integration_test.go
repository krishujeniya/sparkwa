package tests

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/sparkwa/sparkwa/internal/config"
	"github.com/sparkwa/sparkwa/internal/identity"
	"github.com/sparkwa/sparkwa/internal/message"
	"github.com/sparkwa/sparkwa/internal/spark"
	"github.com/sparkwa/sparkwa/internal/store"
	"go.mau.fi/whatsmeow/types"
)

// mockSender records outgoing WhatsApp notifications for integration verification.
type mockSender struct {
	sent []string
}

func (m *mockSender) SendTextMessage(ctx context.Context, recipientPhone, text string) error {
	m.sent = append(m.sent, text)
	return nil
}

func (m *mockSender) IsLoggedIn() bool {
	return true
}

func TestEndToEndBridgeWorkflow(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "e2e_sparkwa.db")
	ctx := context.Background()

	// 1. Initialize store
	s, err := store.Open(ctx, dbPath)
	if err != nil {
		t.Fatalf("failed to open store: %v", err)
	}
	defer s.Close()

	// 2. Setup configuration
	cfg := &config.Config{
		BotPhoneNumber:        "+12345678900",
		AuthorizedPhoneNumber: "+19876543210",
		BridgeGmailAddress:    "bridge@example.com",
		SparkGmailAddress:     "spark@example.com",
		MaxConcurrentJobs:     5,
		MaxMessageLength:      4000,
		GmailPollInterval:    15 * time.Second,
		TaskTimeout:          10 * time.Minute,
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("config validation failed: %v", err)
	}

	// 3. Setup identity verifier (Gate B: WhatsApp sender authorization)
	verifier, err := identity.NewVerifier(cfg.AuthorizedPhoneNumber, cfg.BotPhoneNumber, nil)
	if err != nil {
		t.Fatalf("verifier setup failed: %v", err)
	}

	// Stranger rejection
	if _, err := verifier.Authorize(ctx, identityNormalize("+15550001111")); err == nil {
		t.Fatalf("expected stranger sender to be rejected")
	}

	// Permitted user authorization
	senderNorm, err := verifier.Authorize(ctx, identityNormalize("+1 (987) 654-3210"))
	if err != nil {
		t.Fatalf("authorized user was rejected: %v", err)
	}
	if senderNorm != "+19876543210" {
		t.Errorf("expected normalized phone +19876543210, got %s", senderNorm)
	}

	// 4. Create task in SQLite (Gate C: Command creation)
	taskID := "e2e-task-uuid-1"
	task := &store.Task{
		ID:          taskID,
		SenderPhone: senderNorm,
		WAMessageID: "wa-msg-e2e-100",
		Command:     "What are the scheduled maintenance windows for tomorrow?",
		State:       store.StateQueued,
		CreatedAt:   time.Now().UTC(),
	}
	if err := s.CreateTask(ctx, task); err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	// 5. Format Gmail command envelope for Gemini Spark
	sparkCmd := &spark.Command{
		RequestID:        task.ID,
		AuthorizedSender: task.SenderPhone,
		Timestamp:        task.CreatedAt,
		Content:          task.Command,
	}
	subject, body, err := spark.FormatCommand(sparkCmd)
	if err != nil {
		t.Fatalf("failed to format Spark command: %v", err)
	}
	if !strings.Contains(subject, taskID) || !strings.Contains(body, "BEGIN SPARKWA COMMAND") {
		t.Fatalf("invalid command format: subject=%s, body=%s", subject, body)
	}

	// Simulate successful Gmail dispatch
	_ = s.UpdateGmailIdentifiers(ctx, task.ID, "gmail-msg-e2e", "gmail-thread-e2e")
	_ = s.UpdateTaskState(ctx, task.ID, store.StateAwaitingSpark, "", "")

	// 6. Simulate return response from Gemini Spark (Gate D & Gate E)
	sparkResponseBody := `
Here is the maintenance report from Gemini Spark:

BEGIN SPARKWA RESPONSE
Request ID: ` + taskID + `
Status: SUCCESS
Result: No maintenance windows are scheduled for tomorrow, November 1st.
END SPARKWA RESPONSE
`
	parsed, err := spark.ParseResponse(taskID, sparkResponseBody)
	if err != nil {
		t.Fatalf("failed to parse Spark response: %v", err)
	}
	if parsed.Status != spark.StatusSuccess {
		t.Fatalf("expected StatusSuccess, got %s", parsed.Status)
	}

	// 7. Update task to COMPLETED and forward to WhatsApp mock
	_ = s.UpdateTaskState(ctx, taskID, store.StateCompleted, parsed.ResultText, "")
	waMock := &mockSender{}
	completionMsg := message.FormatCompletion(taskID, parsed.ResultText)
	_ = waMock.SendTextMessage(ctx, senderNorm, completionMsg)

	if len(waMock.sent) != 1 {
		t.Fatalf("expected 1 WhatsApp notification sent, got %d", len(waMock.sent))
	}
	if !strings.Contains(waMock.sent[0], "No maintenance windows are scheduled") {
		t.Errorf("unexpected WhatsApp message content: %s", waMock.sent[0])
	}

	// 8. Verify capability state update
	cm := spark.NewCapabilityManager(s, cfg)
	if err := cm.SetState(ctx, config.CapResponsePathVerified); err != nil {
		t.Fatalf("failed to set capability state: %v", err)
	}
	rep, _ := cm.Evaluate(ctx)
	if rep.State != config.CapResponsePathVerified {
		t.Errorf("expected capability RESPONSE_PATH_VERIFIED, got %s", rep.State)
	}

	// 9. Verify Restart Recovery (Gate F: Reliability)
	// Add an unsubmitted task and test recovery
	_ = s.CreateTask(ctx, &store.Task{
		ID:          "recovered-task-1",
		SenderPhone: senderNorm,
		WAMessageID: "wa-unsubmitted",
		Command:     "Check logs",
		State:       store.StateQueued,
	})
	active, err := s.GetActiveTasks(ctx)
	if err != nil {
		t.Fatalf("GetActiveTasks failed: %v", err)
	}
	if len(active) != 1 || active[0].ID != "recovered-task-1" {
		t.Errorf("expected 1 active unsubmitted task for restart recovery, got %+v", active)
	}

	// 10. Verify Deduplication
	_ = s.MarkMessageProcessed(ctx, "wa-msg-e2e-100", "whatsapp")
	isDup, err := s.IsMessageProcessed(ctx, "wa-msg-e2e-100", "whatsapp")
	if err != nil || !isDup {
		t.Errorf("expected message to be marked duplicate")
	}
}

// helper to build JID for testing verifier
func identityNormalize(phone string) types.JID {
	norm, _ := identity.NormalizeE164(phone)
	digits := strings.TrimPrefix(norm, "+")
	return types.NewJID(digits, types.DefaultUserServer)
}
