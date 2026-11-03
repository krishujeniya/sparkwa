package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

var (
	ErrTaskNotFound = errors.New("task not found")
	ErrDuplicateWA  = errors.New("whatsapp message id already recorded")
)

// Store provides thread-safe persistent SQLite storage for SparkWA.
type Store struct {
	db *sql.DB
	mu sync.RWMutex
}

// Open initializes SQLite storage at the given path, creating parent directories and tables.
func Open(ctx context.Context, dbPath string) (*Store, error) {
	if dbPath != ":memory:" && !filepath.IsAbs(dbPath) {
		abs, err := filepath.Abs(dbPath)
		if err == nil {
			dbPath = abs
		}
	}

	if dbPath != ":memory:" {
		dir := filepath.Dir(dbPath)
		if err := os.MkdirAll(dir, 0700); err != nil {
			return nil, fmt.Errorf("failed to create db directory %s: %w", dir, err)
		}
	}

	dsn := fmt.Sprintf("file:%s?_busy_timeout=30000&_journal_mode=WAL&_foreign_keys=on&_synchronous=NORMAL", dbPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	db.SetMaxOpenConns(1) // SQLite works best with single writer in embedded mode
	db.SetMaxIdleConns(1)

	s := &Store{db: db}
	if err := s.migrate(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("database migration failed: %w", err)
	}

	return s, nil
}

// Close closes the database connection.
func (s *Store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.db.Close()
}

// DB returns the underlying database handle for health checks.
func (s *Store) DB() *sql.DB {
	return s.db
}

func (s *Store) migrate(ctx context.Context) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS tasks (
			id TEXT PRIMARY KEY,
			sender_phone TEXT NOT NULL,
			wa_message_id TEXT NOT NULL UNIQUE,
			command TEXT NOT NULL,
			state TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL,
			submitted_at TIMESTAMP,
			completed_at TIMESTAMP,
			gmail_message_id TEXT,
			gmail_thread_id TEXT,
			response_text TEXT,
			error_message TEXT,
			retry_count INTEGER NOT NULL DEFAULT 0
		);`,
		`CREATE INDEX IF NOT EXISTS idx_tasks_state ON tasks(state);`,
		`CREATE INDEX IF NOT EXISTS idx_tasks_created_at ON tasks(created_at);`,
		`CREATE INDEX IF NOT EXISTS idx_tasks_wa_message_id ON tasks(wa_message_id);`,

		`CREATE TABLE IF NOT EXISTS processed_messages (
			message_id TEXT PRIMARY KEY,
			source TEXT NOT NULL,
			processed_at TIMESTAMP NOT NULL
		);`,
		`CREATE INDEX IF NOT EXISTS idx_processed_messages_source ON processed_messages(source);`,

		`CREATE TABLE IF NOT EXISTS integration_state (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL,
			updated_at TIMESTAMP NOT NULL
		);`,
	}

	for _, q := range queries {
		if _, err := s.db.ExecContext(ctx, q); err != nil {
			return fmt.Errorf("executing migration query failed: %w (query: %s)", err, q)
		}
	}

	return nil
}

// CreateTask saves a new task to SQLite.
func (s *Store) CreateTask(ctx context.Context, t *Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	if t.CreatedAt.IsZero() {
		t.CreatedAt = now
	}
	t.UpdatedAt = now

	query := `INSERT INTO tasks (
		id, sender_phone, wa_message_id, command, state,
		created_at, updated_at, submitted_at, completed_at,
		gmail_message_id, gmail_thread_id, response_text, error_message, retry_count
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := s.db.ExecContext(ctx, query,
		t.ID, t.SenderPhone, t.WAMessageID, t.Command, string(t.State),
		t.CreatedAt, t.UpdatedAt, t.SubmittedAt, t.CompletedAt,
		t.GmailMessageID, t.GmailThreadID, t.ResponseText, t.ErrorMessage, t.RetryCount,
	)
	if err != nil {
		return err
	}
	return nil
}

// GetTask retrieves a task by its unique ID.
func (s *Store) GetTask(ctx context.Context, id string) (*Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `SELECT 
		id, sender_phone, wa_message_id, command, state,
		created_at, updated_at, submitted_at, completed_at,
		gmail_message_id, gmail_thread_id, response_text, error_message, retry_count
	FROM tasks WHERE id = ?`

	row := s.db.QueryRowContext(ctx, query, id)
	return scanTask(row)
}

// GetTaskByWAMessageID retrieves a task by the incoming WhatsApp message ID for deduplication.
func (s *Store) GetTaskByWAMessageID(ctx context.Context, waMsgID string) (*Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `SELECT 
		id, sender_phone, wa_message_id, command, state,
		created_at, updated_at, submitted_at, completed_at,
		gmail_message_id, gmail_thread_id, response_text, error_message, retry_count
	FROM tasks WHERE wa_message_id = ?`

	row := s.db.QueryRowContext(ctx, query, waMsgID)
	return scanTask(row)
}

// UpdateTaskState updates the lifecycle state, optional fields, and timestamp.
func (s *Store) UpdateTaskState(ctx context.Context, id string, state TaskState, responseText, errorMsg string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	var query string
	var args []any

	if state == StateCompleted {
		query = `UPDATE tasks SET state = ?, response_text = ?, error_message = ?, updated_at = ?, completed_at = ? WHERE id = ?`
		args = []any{string(state), responseText, errorMsg, now, now, id}
	} else if state == StateSubmitted || state == StateAwaitingSpark {
		query = `UPDATE tasks SET state = ?, response_text = ?, error_message = ?, updated_at = ?, submitted_at = ? WHERE id = ?`
		args = []any{string(state), responseText, errorMsg, now, now, id}
	} else {
		query = `UPDATE tasks SET state = ?, response_text = ?, error_message = ?, updated_at = ? WHERE id = ?`
		args = []any{string(state), responseText, errorMsg, now, id}
	}

	res, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrTaskNotFound
	}
	return nil
}

// UpdateGmailIdentifiers sets the sent Gmail Message ID and Thread ID.
func (s *Store) UpdateGmailIdentifiers(ctx context.Context, id, gmailMsgID, threadID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `UPDATE tasks SET gmail_message_id = ?, gmail_thread_id = ?, updated_at = ? WHERE id = ?`
	_, err := s.db.ExecContext(ctx, query, gmailMsgID, threadID, time.Now().UTC(), id)
	return err
}

// IncrementRetry increments the task's retry counter and updates state.
func (s *Store) IncrementRetry(ctx context.Context, id string, errorMsg string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `UPDATE tasks SET retry_count = retry_count + 1, error_message = ?, updated_at = ? WHERE id = ?`
	_, err := s.db.ExecContext(ctx, query, errorMsg, time.Now().UTC(), id)
	return err
}

// ListRecentTasks returns the N most recently created tasks.
func (s *Store) ListRecentTasks(ctx context.Context, limit int) ([]*Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `SELECT 
		id, sender_phone, wa_message_id, command, state,
		created_at, updated_at, submitted_at, completed_at,
		gmail_message_id, gmail_thread_id, response_text, error_message, retry_count
	FROM tasks ORDER BY created_at DESC LIMIT ?`

	rows, err := s.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		t, err := scanTaskRow(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

// GetActiveTasks returns all non-terminal tasks (for crash recovery on restart).
func (s *Store) GetActiveTasks(ctx context.Context) ([]*Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `SELECT 
		id, sender_phone, wa_message_id, command, state,
		created_at, updated_at, submitted_at, completed_at,
		gmail_message_id, gmail_thread_id, response_text, error_message, retry_count
	FROM tasks WHERE state IN (?, ?, ?, ?, ?) ORDER BY created_at ASC`

	rows, err := s.db.QueryContext(ctx, query,
		string(StateReceived),
		string(StateAuthorized),
		string(StateQueued),
		string(StateSubmitted),
		string(StateAwaitingSpark),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		t, err := scanTaskRow(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

// ExpireOldTasks marks any active tasks older than the timeout as EXPIRED.
func (s *Store) ExpireOldTasks(ctx context.Context, timeout time.Duration) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	threshold := time.Now().UTC().Add(-timeout)
	query := `UPDATE tasks 
		SET state = ?, error_message = 'Task timed out waiting for Spark response', updated_at = ? 
		WHERE state IN (?, ?, ?, ?, ?) AND created_at < ?`

	res, err := s.db.ExecContext(ctx, query,
		string(StateExpired),
		time.Now().UTC(),
		string(StateReceived),
		string(StateAuthorized),
		string(StateQueued),
		string(StateSubmitted),
		string(StateAwaitingSpark),
		threshold,
	)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// CancelTask marks a task as CANCELLED if it has not completed.
func (s *Store) CancelTask(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	query := `UPDATE tasks SET state = ?, updated_at = ? 
		WHERE id = ? AND state NOT IN (?, ?, ?, ?)`

	res, err := s.db.ExecContext(ctx, query,
		string(StateCancelled), now, id,
		string(StateCompleted), string(StateFailed), string(StateExpired), string(StateCancelled),
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return errors.New("task cannot be cancelled (either not found or already completed/failed)")
	}
	return nil
}

// IsMessageProcessed checks if an inbound message has already been processed.
func (s *Store) IsMessageProcessed(ctx context.Context, messageID, source string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var exists int
	query := `SELECT 1 FROM processed_messages WHERE message_id = ? AND source = ? LIMIT 1`
	err := s.db.QueryRowContext(ctx, query, messageID, source).Scan(&exists)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// MarkMessageProcessed records an inbound message ID to prevent duplicate handling.
func (s *Store) MarkMessageProcessed(ctx context.Context, messageID, source string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `INSERT OR IGNORE INTO processed_messages (message_id, source, processed_at) VALUES (?, ?, ?)`
	_, err := s.db.ExecContext(ctx, query, messageID, source, time.Now().UTC())
	return err
}

// GetIntegrationState retrieves a persistent key-value configuration or capability state.
func (s *Store) GetIntegrationState(ctx context.Context, key string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var val string
	query := `SELECT value FROM integration_state WHERE key = ?`
	err := s.db.QueryRowContext(ctx, query, key).Scan(&val)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return val, err
}

// SetIntegrationState persists a key-value capability or status indicator.
func (s *Store) SetIntegrationState(ctx context.Context, key, val string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `INSERT INTO integration_state (key, value, updated_at) 
		VALUES (?, ?, ?) 
		ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = excluded.updated_at`
	_, err := s.db.ExecContext(ctx, query, key, val, time.Now().UTC())
	return err
}

type scanner interface {
	Scan(dest ...any) error
}

func scanTask(row scanner) (*Task, error) {
	var t Task
	var st string
	var subAt, compAt sql.NullTime
	var gMsgID, gThreadID, respText, errMsg sql.NullString

	err := row.Scan(
		&t.ID, &t.SenderPhone, &t.WAMessageID, &t.Command, &st,
		&t.CreatedAt, &t.UpdatedAt, &subAt, &compAt,
		&gMsgID, &gThreadID, &respText, &errMsg, &t.RetryCount,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrTaskNotFound
	}
	if err != nil {
		return nil, err
	}

	t.State = TaskState(st)
	if subAt.Valid {
		t.SubmittedAt = &subAt.Time
	}
	if compAt.Valid {
		t.CompletedAt = &compAt.Time
	}
	if gMsgID.Valid {
		t.GmailMessageID = gMsgID.String
	}
	if gThreadID.Valid {
		t.GmailThreadID = gThreadID.String
	}
	if respText.Valid {
		t.ResponseText = respText.String
	}
	if errMsg.Valid {
		t.ErrorMessage = errMsg.String
	}

	return &t, nil
}

func scanTaskRow(rows *sql.Rows) (*Task, error) {
	return scanTask(rows)
}
