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

func TestStoreListRecentAndRetry(t *testing.T) {
	s := setupTestStore(t)
	ctx := context.Background()

	// List on empty store
	emptyList, err := s.ListRecentTasks(ctx, 10)
	if err != nil {
		t.Fatalf("ListRecentTasks on empty store failed: %v", err)
	}
	if len(emptyList) != 0 {
		t.Fatalf("expected 0 tasks, got %d", len(emptyList))
	}

	task1 := &Task{ID: "t-1", SenderPhone: "+100", WAMessageID: "m1", Command: "c1", State: StateQueued}
	task2 := &Task{ID: "t-2", SenderPhone: "+100", WAMessageID: "m2", Command: "c2", State: StateQueued}
	_ = s.CreateTask(ctx, task1)
	_ = s.CreateTask(ctx, task2)

	recent, err := s.ListRecentTasks(ctx, 1)
	if err != nil {
		t.Fatalf("ListRecentTasks failed: %v", err)
	}
	if len(recent) != 1 {
		t.Fatalf("expected 1 task with limit 1, got %d", len(recent))
	}

	// Test increment retry
	if err := s.IncrementRetry(ctx, "t-1", "temporary network timeout"); err != nil {
		t.Fatalf("IncrementRetry failed: %v", err)
	}
	t1, err := s.GetTask(ctx, "t-1")
	if err != nil {
		t.Fatalf("GetTask failed: %v", err)
	}
	if t1.RetryCount != 1 || t1.ErrorMessage != "temporary network timeout" {
		t.Errorf("unexpected retry state: count=%d, err=%s", t1.RetryCount, t1.ErrorMessage)
	}

	// Test non-existent task lookups
	_, err = s.GetTask(ctx, "non-existent-id")
	if err != ErrTaskNotFound {
		t.Errorf("expected ErrTaskNotFound, got %v", err)
	}

	err = s.UpdateTaskState(ctx, "non-existent-id", StateCompleted, "", "")
	if err != ErrTaskNotFound {
		t.Errorf("expected ErrTaskNotFound on update, got %v", err)
	}
}

func BenchmarkStoreTaskCreation(b *testing.B) {
	tempDir := b.TempDir()
	dbPath := filepath.Join(tempDir, "bench_sparkwa.db")
	s, err := Open(context.Background(), dbPath)
	if err != nil {
		b.Fatalf("failed to open bench store: %v", err)
	}
	defer s.Close()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		t := &Task{
			ID:          filepath.Base(tempDir) + string(rune(i)),
			SenderPhone: "+19876543210",
			WAMessageID: "wa-" + string(rune(i)),
			Command:     "Benchmark test command",
			State:       StateQueued,
		}
		_ = s.CreateTask(ctx, t)
	}
}
