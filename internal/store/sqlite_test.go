package store

import (
	"context"
	"path/filepath"
	"testing"
	"time"
)

func setupTestStore(t *testing.T) *Store {
	t.Helper()
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_sparkwa.db")

	s, err := Open(context.Background(), dbPath)
	if err != nil {
		t.Fatalf("failed to open test store: %v", err)
	}
	t.Cleanup(func() {
		_ = s.Close()
	})
	return s
}

func TestStoreTaskLifecycle(t *testing.T) {
	s := setupTestStore(t)
	ctx := context.Background()

	task := &Task{
		ID:          "task-1234-uuid",
		SenderPhone: "+19876543210",
		WAMessageID: "wa-msg-999",
		Command:     "What are my scheduled meetings for today?",
		State:       StateQueued,
	}

	// 1. Create Task
	if err := s.CreateTask(ctx, task); err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	// 2. Fetch by ID
	fetched, err := s.GetTask(ctx, task.ID)
	if err != nil {
		t.Fatalf("GetTask failed: %v", err)
	}
	if fetched.Command != task.Command || fetched.State != StateQueued {
		t.Errorf("task data mismatch: got %+v, want %+v", fetched, task)
	}

	// 3. Fetch by WhatsApp message ID (for deduplication)
	byWAMsg, err := s.GetTaskByWAMessageID(ctx, "wa-msg-999")
	if err != nil {
		t.Fatalf("GetTaskByWAMessageID failed: %v", err)
	}
	if byWAMsg.ID != task.ID {
		t.Errorf("ID mismatch: got %s, want %s", byWAMsg.ID, task.ID)
	}

	// 4. Update to SUBMITTED
	if err := s.UpdateTaskState(ctx, task.ID, StateSubmitted, "", ""); err != nil {
		t.Fatalf("UpdateTaskState SUBMITTED failed: %v", err)
	}
	_ = s.UpdateGmailIdentifiers(ctx, task.ID, "gmail-msg-001", "gmail-thread-001")

	// 5. Update to COMPLETED
	if err := s.UpdateTaskState(ctx, task.ID, StateCompleted, "You have 2 meetings today.", ""); err != nil {
		t.Fatalf("UpdateTaskState COMPLETED failed: %v", err)
	}

	completed, err := s.GetTask(ctx, task.ID)
	if err != nil {
		t.Fatalf("GetTask completed failed: %v", err)
	}
	if completed.State != StateCompleted || completed.ResponseText != "You have 2 meetings today." {
		t.Errorf("unexpected completed state: %+v", completed)
	}
	if completed.GmailMessageID != "gmail-msg-001" {
		t.Errorf("expected gmail message id, got %s", completed.GmailMessageID)
	}
}

func TestStoreActiveTasksRecovery(t *testing.T) {
	s := setupTestStore(t)
	ctx := context.Background()

	// Insert 3 tasks: 1 QUEUED, 1 AWAITING_SPARK, 1 COMPLETED
	_ = s.CreateTask(ctx, &Task{ID: "t1", SenderPhone: "+100", WAMessageID: "m1", Command: "c1", State: StateQueued})
	_ = s.CreateTask(ctx, &Task{ID: "t2", SenderPhone: "+100", WAMessageID: "m2", Command: "c2", State: StateAwaitingSpark})
	_ = s.CreateTask(ctx, &Task{ID: "t3", SenderPhone: "+100", WAMessageID: "m3", Command: "c3", State: StateCompleted})

	active, err := s.GetActiveTasks(ctx)
	if err != nil {
		t.Fatalf("GetActiveTasks failed: %v", err)
	}
	if len(active) != 2 {
		t.Fatalf("expected 2 active tasks, got %d", len(active))
	}
	if active[0].ID != "t1" || active[1].ID != "t2" {
		t.Errorf("unexpected active tasks order or ids: %+v", active)
	}
}

func TestStoreTaskExpiration(t *testing.T) {
	s := setupTestStore(t)
	ctx := context.Background()

	oldTime := time.Now().UTC().Add(-30 * time.Minute)
	_ = s.CreateTask(ctx, &Task{
		ID:          "expired-task",
		SenderPhone: "+100",
		WAMessageID: "m-old",
		Command:     "old command",
		State:       StateAwaitingSpark,
		CreatedAt:   oldTime,
	})

	count, err := s.ExpireOldTasks(ctx, 15*time.Minute)
	if err != nil {
		t.Fatalf("ExpireOldTasks failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 task expired, got %d", count)
	}

	expired, err := s.GetTask(ctx, "expired-task")
	if err != nil {
		t.Fatalf("GetTask failed: %v", err)
	}
	if expired.State != StateExpired {
		t.Errorf("expected state EXPIRED, got %s", expired.State)
	}
}

func TestStoreCancellation(t *testing.T) {
	s := setupTestStore(t)
	ctx := context.Background()

	_ = s.CreateTask(ctx, &Task{ID: "t-cancel", SenderPhone: "+100", WAMessageID: "m-c", Command: "test", State: StateQueued})

	if err := s.CancelTask(ctx, "t-cancel"); err != nil {
		t.Fatalf("CancelTask failed: %v", err)
	}

	cancelled, _ := s.GetTask(ctx, "t-cancel")
	if cancelled.State != StateCancelled {
		t.Errorf("expected StateCancelled, got %s", cancelled.State)
	}

	// Trying to cancel an already cancelled or completed task should return error
	if err := s.CancelTask(ctx, "t-cancel"); err == nil {
		t.Errorf("expected error when cancelling already cancelled task, got nil")
	}
}

func TestStoreMessageDeduplication(t *testing.T) {
	s := setupTestStore(t)
	ctx := context.Background()

	processed, err := s.IsMessageProcessed(ctx, "wa-id-123", "whatsapp")
	if err != nil || processed {
		t.Fatalf("message should not be processed yet")
	}

	if err := s.MarkMessageProcessed(ctx, "wa-id-123", "whatsapp"); err != nil {
		t.Fatalf("MarkMessageProcessed failed: %v", err)
	}

	processed, err = s.IsMessageProcessed(ctx, "wa-id-123", "whatsapp")
	if err != nil || !processed {
		t.Fatalf("message should now be reported as processed")
	}
}

func TestStoreIntegrationState(t *testing.T) {
	s := setupTestStore(t)
	ctx := context.Background()

	if err := s.SetIntegrationState(ctx, "capability_state", "RESPONSE_PATH_VERIFIED"); err != nil {
		t.Fatalf("SetIntegrationState failed: %v", err)
	}

	val, err := s.GetIntegrationState(ctx, "capability_state")
	if err != nil {
		t.Fatalf("GetIntegrationState failed: %v", err)
	}
	if val != "RESPONSE_PATH_VERIFIED" {
		t.Errorf("got %q, want RESPONSE_PATH_VERIFIED", val)
	}
}

func TestTaskFlowTag(t *testing.T) {
	tests := []struct {
		state TaskState
		want  string
	}{
		{StateReceived, "2 -> 1"},
		{StateAuthorized, "2 -> 1"},
		{StateQueued, "2 -> 1"},
		{StateSubmitted, "1 -> E"},
		{StateAwaitingSpark, "E -> S"},
		{StateApprovalRequired, "S -> 2"},
		{StateCompleted, "E -> 2"},
		{StateFailed, "! -> 2"},
		{StateExpired, "! -> 2"},
		{StateCancelled, "! -> 2"},
	}

	for _, tc := range tests {
		task := &Task{State: tc.state}
		got := task.FlowTag()
		if got != tc.want {
			t.Errorf("FlowTag() for state %s = %s, want %s", tc.state, got, tc.want)
		}
	}
}
