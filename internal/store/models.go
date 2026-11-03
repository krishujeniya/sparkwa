package store

import (
	"time"
)

// TaskState represents the lifecycle status of a user command.
type TaskState string

const (
	StateReceived         TaskState = "RECEIVED"
	StateAuthorized       TaskState = "AUTHORIZED"
	StateQueued           TaskState = "QUEUED"
	StateSubmitted        TaskState = "SUBMITTED"
	StateAwaitingSpark    TaskState = "AWAITING_SPARK"
	StateApprovalRequired TaskState = "APPROVAL_REQUIRED"
	StateCompleted        TaskState = "COMPLETED"
	StateFailed           TaskState = "FAILED"
	StateExpired          TaskState = "EXPIRED"
	StateCancelled        TaskState = "CANCELLED"
)

// Task represents a persistent unit of work bridging WhatsApp and Gemini Spark.
type Task struct {
	ID             string     `json:"id"`
	SenderPhone    string     `json:"sender_phone"`
	WAMessageID    string     `json:"wa_message_id"`
	Command        string     `json:"command"`
	State          TaskState  `json:"state"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	SubmittedAt    *time.Time `json:"submitted_at,omitempty"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
	GmailMessageID string     `json:"gmail_message_id,omitempty"`
	GmailThreadID  string     `json:"gmail_thread_id,omitempty"`
	ResponseText   string     `json:"response_text,omitempty"`
	ErrorMessage   string     `json:"error_message,omitempty"`
	RetryCount     int        `json:"retry_count"`
}

// IsTerminal returns true if the task reached a terminal state.
func (t *Task) IsTerminal() bool {
	switch t.State {
	case StateCompleted, StateFailed, StateExpired, StateCancelled:
		return true
	default:
		return false
	}
}

// CanCancel returns true if the task can still be safely cancelled.
func (t *Task) CanCancel() bool {
	switch t.State {
	case StateReceived, StateAuthorized, StateQueued:
		return true
	case StateSubmitted, StateAwaitingSpark:
		// Best-effort cancellation after reaching Spark
		return true
	default:
		return false
	}
}

// FlowTag returns a privacy-safe node flow indicator for the task's pipeline progress.
// Node [1]: WhatsApp Bot
// Node [2]: Authorized User
// Node [E]: Gmail Bridge
// Node [S]: Gemini Spark Agent
func (t *Task) FlowTag() string {
	switch t.State {
	case StateReceived, StateAuthorized, StateQueued:
		return "2 -> 1"
	case StateSubmitted:
		return "1 -> E"
	case StateAwaitingSpark:
		return "E -> S"
	case StateApprovalRequired:
		return "S -> 2"
	case StateCompleted:
		return "E -> 2"
	case StateFailed, StateExpired, StateCancelled:
		return "! -> 2"
	default:
		return "1 -> E"
	}
}

