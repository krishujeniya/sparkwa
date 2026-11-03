package message

import (
	"strings"
	"testing"
	"time"

	"github.com/sparkwa/sparkwa/internal/config"
	"github.com/sparkwa/sparkwa/internal/store"
)

func TestFormatApprovalNeeded(t *testing.T) {
	prompt := "Please confirm deleting the old invoice file."
	out := FormatApprovalNeeded("task-123", prompt)

	if !strings.Contains(out, "[APPROVAL REQUIRED]") {
		t.Errorf("expected [APPROVAL REQUIRED] in output, got: %s", out)
	}
	if !strings.Contains(out, prompt) {
		t.Errorf("expected prompt text in output, got: %s", out)
	}
}

func TestFormatFailure(t *testing.T) {
	errText := "Gmail API returned quota exceeded"
	out := FormatFailure("task-456", errText)

	if !strings.Contains(out, "[ERROR]") || !strings.Contains(out, errText) {
		t.Errorf("unexpected failure format: %s", out)
	}
}

func TestFormatStatusReport(t *testing.T) {
	out := FormatStatusReport(
		true,
		"+1234567890",
		"+1987654321",
		true,
		"bridge@example.com",
		"spark@example.com",
		config.CapResponsePathVerified,
		3,
	)

	if !strings.Contains(out, "*SparkWA Status*") {
		t.Errorf("missing status header: %s", out)
	}
	if !strings.Contains(out, "[CONNECTED]") {
		t.Errorf("missing WhatsApp connected indicator: %s", out)
	}
	if !strings.Contains(out, "[ACTIVE]") {
		t.Errorf("missing Gmail active indicator: %s", out)
	}
	if !strings.Contains(out, "Active Jobs in Queue:* 3") {
		t.Errorf("missing job queue count: %s", out)
	}

	// Disconnected test
	outDisc := FormatStatusReport(
		false,
		"+1234567890",
		"+1987654321",
		false,
		"bridge@example.com",
		"spark@example.com",
		config.CapNotConfigured,
		0,
	)
	if !strings.Contains(outDisc, "[DISCONNECTED]") {
		t.Errorf("expected disconnected status: %s", outDisc)
	}
}

func TestFormatJobList(t *testing.T) {
	empty := FormatJobList(nil)
	if !strings.Contains(empty, "No recent tasks recorded") {
		t.Errorf("expected empty task notice: %s", empty)
	}

	now := time.Now().UTC()
	tasks := []*store.Task{
		{
			ID:        "abcd1234efgh",
			State:     store.StateCompleted,
			Command:   "Summarize Q3 earnings report for team",
			CreatedAt: now,
		},
		{
			ID:        "deadbeef9999",
			State:     store.StateFailed,
			Command:   "Execute invalid query with syntax error",
			CreatedAt: now,
		},
		{
			ID:        "feedface1111",
			State:     store.StateApprovalRequired,
			Command:   "Delete legacy client folder",
			CreatedAt: now,
		},
		{
			ID:        "cafeed002222",
			State:     store.StateCancelled,
			Command:   "Cancelled command task",
			CreatedAt: now,
		},
		{
			ID:        "baadf00d3333",
			State:     store.StateSubmitted,
			Command:   "Running task now",
			CreatedAt: now,
		},
	}

	out := FormatJobList(tasks)
	if !strings.Contains(out, "[DONE]") {
		t.Errorf("expected [DONE] for completed task: %s", out)
	}
	if !strings.Contains(out, "[FAIL]") {
		t.Errorf("expected [FAIL] for failed task: %s", out)
	}
	if !strings.Contains(out, "[APPR]") {
		t.Errorf("expected [APPR] for approval task: %s", out)
	}
	if !strings.Contains(out, "[CANC]") {
		t.Errorf("expected [CANC] for cancelled task: %s", out)
	}
	if !strings.Contains(out, "[RUN]") {
		t.Errorf("expected [RUN] for submitted task: %s", out)
	}
}

func TestFormatJobDetail(t *testing.T) {
	now := time.Now().UTC()
	sub := now.Add(1 * time.Second)
	comp := now.Add(5 * time.Second)

	task := &store.Task{
		ID:             "256c0c48",
		State:          store.StateCompleted,
		Command:        "List calendar events for today",
		ResponseText:   "You have 3 meetings scheduled today.",
		GmailMessageID: "msg-gmail-1234",
		RetryCount:     1,
		CreatedAt:      now,
		SubmittedAt:    &sub,
		CompletedAt:    &comp,
	}

	detail := FormatJobDetail(task)
	if !strings.Contains(detail, "*Task Detail:* `256c0c48`") {
		t.Errorf("missing task ID: %s", detail)
	}
	if !strings.Contains(detail, "List calendar events for today") {
		t.Errorf("missing original command: %s", detail)
	}
	if !strings.Contains(detail, "You have 3 meetings scheduled today.") {
		t.Errorf("missing response text: %s", detail)
	}
	if !strings.Contains(detail, "msg-gmail-1234") {
		t.Errorf("missing Gmail message ID: %s", detail)
	}
}

func BenchmarkFormatStatusReport(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = FormatStatusReport(
			true,
			"+1234567890",
			"+1987654321",
			true,
			"bridge@example.com",
			"spark@example.com",
			config.CapResponsePathVerified,
			5,
		)
	}
}

func BenchmarkFormatJobDetail(b *testing.B) {
	now := time.Now().UTC()
	sub := now.Add(1 * time.Second)
	comp := now.Add(5 * time.Second)
	task := &store.Task{
		ID:             "256c0c48",
		State:          store.StateCompleted,
		Command:        "List calendar events for today",
		ResponseText:   "You have 3 meetings scheduled today.",
		GmailMessageID: "msg-gmail-1234",
		RetryCount:     1,
		CreatedAt:      now,
		SubmittedAt:    &sub,
		CompletedAt:    &comp,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = FormatJobDetail(task)
	}
}
