package worker

import (
	"context"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/sparkwa/sparkwa-lite/internal/config"
	"github.com/sparkwa/sparkwa-lite/internal/gmail"
	"github.com/sparkwa/sparkwa-lite/internal/spark"
	"github.com/sparkwa/sparkwa-lite/internal/store"
)

// MockWhatsAppSender implements WhatsAppSender for unit tests.
type mockWASender struct {
	mu       sync.Mutex
	messages []sentRecord
	loggedIn bool
}

type sentRecord struct {
	Recipient string
	Text      string
}

func (m *mockWASender) SendTextMessage(ctx context.Context, recipientPhone, text string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = append(m.messages, sentRecord{Recipient: recipientPhone, Text: text})
	return nil
}

func (m *mockWASender) IsLoggedIn() bool {
	return m.loggedIn
}

func (m *mockWASender) lastMessage() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.messages) == 0 {
		return ""
	}
	return m.messages[len(m.messages)-1].Text
}

func setupTestEngine(t *testing.T) (*Engine, *store.Store, *mockWASender) {
	t.Helper()
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_engine.db")

	s, err := store.Open(context.Background(), dbPath)
	if err != nil {
		t.Fatalf("failed to open test store: %v", err)
	}
	t.Cleanup(func() {
		_ = s.Close()
	})

	cfg := &config.Config{
		BotPhoneNumber:        "+12345678900",
		AuthorizedPhoneNumber: "+19876543210",
		BridgeGmailAddress:    "bridge@example.com",
		SparkGmailAddress:     "spark@example.com",
		MaxConcurrentJobs:     5,
		MaxMessageLength:      4000,
		GmailPollInterval:    10 * time.Second,
		TaskTimeout:          15 * time.Minute,
	}

	wa := &mockWASender{loggedIn: true}
	cm := spark.NewCapabilityManager(s, cfg)

	eng := NewEngine(cfg, s, nil, nil, wa, cm)
	return eng, s, wa
}

func TestEngineSlashCommands(t *testing.T) {
	eng, s, wa := setupTestEngine(t)
	ctx := context.Background()

	// 1. /help
	eng.HandleInboundWhatsApp(ctx, "wa-1", "+19876543210", "/help")
	last := wa.lastMessage()
	if !strings.Contains(last, "/status") || !strings.Contains(last, "Available Commands") {
		t.Errorf("unexpected /help reply: %s", last)
	}

	// 2. /ping
	eng.HandleInboundWhatsApp(ctx, "wa-2", "+19876543210", "/ping")
	last = wa.lastMessage()
	if !strings.Contains(last, "Pong") {
		t.Errorf("unexpected /ping reply: %s", last)
	}

	// 3. /status
	eng.HandleInboundWhatsApp(ctx, "wa-3", "+19876543210", "/status")
	last = wa.lastMessage()
	if !strings.Contains(last, "SparkWA Lite Status") {
		t.Errorf("unexpected /status reply: %s", last)
	}

	// 4. /jobs (empty)
	eng.HandleInboundWhatsApp(ctx, "wa-4", "+19876543210", "/jobs")
	last = wa.lastMessage()
	if !strings.Contains(last, "No recent tasks") {
		t.Errorf("unexpected /jobs empty reply: %s", last)
	}

	// Add a dummy task and test /jobs and /job ID and /cancel ID
	_ = s.CreateTask(ctx, &store.Task{
		ID:          "uuid-task-9999",
		SenderPhone: "+19876543210",
		WAMessageID: "wa-orig",
		Command:     "Test task command",
		State:       store.StateQueued,
	})

	eng.HandleInboundWhatsApp(ctx, "wa-5", "+19876543210", "/jobs")
	last = wa.lastMessage()
	if !strings.Contains(last, "uuid-tas") {
		t.Errorf("expected task listed in /jobs: %s", last)
	}

	eng.HandleInboundWhatsApp(ctx, "wa-6", "+19876543210", "/job uuid-task-9999")
	last = wa.lastMessage()
	if !strings.Contains(last, "Task Detail") || !strings.Contains(last, "Test task command") {
		t.Errorf("unexpected /job detail: %s", last)
	}

	eng.HandleInboundWhatsApp(ctx, "wa-7", "+19876543210", "/cancel uuid-task-9999")
	last = wa.lastMessage()
	if !strings.Contains(last, "has been cancelled") {
		t.Errorf("unexpected /cancel reply: %s", last)
	}
}

func TestEngineMessageDeduplication(t *testing.T) {
	eng, _, wa := setupTestEngine(t)
	ctx := context.Background()

	// Send message 1
	eng.HandleInboundWhatsApp(ctx, "duplicate-wa-id", "+19876543210", "/ping")
	count1 := len(wa.messages)
	if count1 != 1 {
		t.Fatalf("expected 1 message sent, got %d", count1)
	}

	// Send same WhatsApp message ID again
	eng.HandleInboundWhatsApp(ctx, "duplicate-wa-id", "+19876543210", "/ping")
	count2 := len(wa.messages)
	if count2 != 1 {
		t.Fatalf("expected duplicate message to be ignored, but got %d messages", count2)
	}
}

func TestEngineProcessDiscoveredReply(t *testing.T) {
	eng, s, wa := setupTestEngine(t)
	ctx := context.Background()

	taskID := "task-reply-test-1"
	_ = s.CreateTask(ctx, &store.Task{
		ID:          taskID,
		SenderPhone: "+19876543210",
		WAMessageID: "wa-init-1",
		Command:     "Summarize project status",
		State:       store.StateAwaitingSpark,
	})

	// 1. Process successful reply from Spark
	reply := &gmail.DiscoveredReply{
		RequestID:      taskID,
		GmailMessageID: "g-msg-123",
		BodyText: `BEGIN SPARKWA RESPONSE
Request ID: ` + taskID + `
Status: SUCCESS
Result: Project is on track. 10 out of 10 milestones completed.
END SPARKWA RESPONSE`,
	}

	eng.processDiscoveredReply(ctx, reply)

	last := wa.lastMessage()
	if !strings.Contains(last, "COMPLETED") || !strings.Contains(last, "10 out of 10 milestones") {
		t.Errorf("unexpected completion notification to WhatsApp: %s", last)
	}

	updated, _ := s.GetTask(ctx, taskID)
	if updated.State != store.StateCompleted {
		t.Errorf("expected task state COMPLETED, got %s", updated.State)
	}

	// Verify capability updated
	capState, _ := s.GetIntegrationState(ctx, spark.StateKeyCapability)
	if capState != string(config.CapResponsePathVerified) {
		t.Errorf("expected capability RESPONSE_PATH_VERIFIED, got %s", capState)
	}
}

func TestEngineProcessApprovalRequired(t *testing.T) {
	eng, s, wa := setupTestEngine(t)
	ctx := context.Background()

	taskID := "task-approval-test"
	_ = s.CreateTask(ctx, &store.Task{
		ID:          taskID,
		SenderPhone: "+19876543210",
		WAMessageID: "wa-init-2",
		Command:     "Delete old files",
		State:       store.StateAwaitingSpark,
	})

	reply := &gmail.DiscoveredReply{
		RequestID:      taskID,
		GmailMessageID: "g-msg-456",
		BodyText: `BEGIN SPARKWA RESPONSE
Request ID: ` + taskID + `
Status: APPROVAL_REQUIRED
Result: This will delete 45 files permanently. Do you want me to proceed?
END SPARKWA RESPONSE`,
	}

	eng.processDiscoveredReply(ctx, reply)

	last := wa.lastMessage()
	if !strings.Contains(last, "APPROVAL NEEDED") || !strings.Contains(last, "delete 45 files permanently") {
		t.Errorf("unexpected approval prompt sent to WhatsApp: %s", last)
	}

	updated, _ := s.GetTask(ctx, taskID)
	if updated.State != store.StateApprovalRequired {
		t.Errorf("expected task state APPROVAL_REQUIRED, got %s", updated.State)
	}
}
