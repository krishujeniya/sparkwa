package worker

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/sparkwa/sparkwa/internal/config"
	"github.com/sparkwa/sparkwa/internal/gmail"
	"github.com/sparkwa/sparkwa/internal/message"
	"github.com/sparkwa/sparkwa/internal/spark"
	"github.com/sparkwa/sparkwa/internal/store"
)

// WhatsAppSender represents the interface needed to send WhatsApp messages back to the authorized user.
type WhatsAppSender interface {
	SendTextMessage(ctx context.Context, recipientPhone, text string) error
	IsLoggedIn() bool
}

// Engine coordinates message reception, task persistence, Gmail dispatch, and response forwarding.
type Engine struct {
	config     *config.Config
	store      *store.Store
	gmailClient *gmail.Client
	gmailPoller *gmail.Poller
	waSender   WhatsAppSender
	capManager *spark.CapabilityManager
	mu         sync.Mutex
	running    bool
	stopChan   chan struct{}
}

// NewEngine creates a new operational Engine instance.
func NewEngine(
	cfg *config.Config,
	s *store.Store,
	gClient *gmail.Client,
	gPoller *gmail.Poller,
	waSender WhatsAppSender,
	capManager *spark.CapabilityManager,
) *Engine {
	return &Engine{
		config:      cfg,
		store:       s,
		gmailClient: gClient,
		gmailPoller: gPoller,
		waSender:    waSender,
		capManager:  capManager,
		stopChan:    make(chan struct{}),
	}
}

// Start launches the background worker routines (restart recovery, response poller, task sweeper).
func (e *Engine) Start(ctx context.Context) error {
	e.mu.Lock()
	if e.running {
		e.mu.Unlock()
		return nil
	}
	e.running = true
	e.mu.Unlock()

	// 1. Recover active tasks from previous run
	e.recoverActiveTasks(ctx)

	// 2. Launch polling loop
	go e.pollingLoop(ctx)

	// 3. Launch timeout sweeper loop
	go e.sweeperLoop(ctx)

	return nil
}

// Stop gracefully stops the background worker loops.
func (e *Engine) Stop() {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.running {
		return
	}
	close(e.stopChan)
	e.running = false
}

// HandleInboundWhatsApp is called whenever an authorized text message is received from Number 2.
func (e *Engine) HandleInboundWhatsApp(ctx context.Context, msgID, senderPhone, text string) {
	// Deduplicate incoming WhatsApp message ID
	alreadyProcessed, err := e.store.IsMessageProcessed(ctx, msgID, "whatsapp")
	if err != nil {
		log.Printf("[Worker] Error checking deduplication for msg %s: %v", msgID, err)
		return
	}
	if alreadyProcessed {
		log.Printf("[Worker] Duplicate WhatsApp message %s ignored", msgID)
		return
	}
	_ = e.store.MarkMessageProcessed(ctx, msgID, "whatsapp")

	trimmed := strings.TrimSpace(text)
	log.Printf("[Worker] Inbound message accepted from %s (msgID: %s): %q", senderPhone, msgID, trimmed)

	// Check for Slash Commands
	if strings.HasPrefix(trimmed, "/") {
		e.handleSlashCommand(ctx, senderPhone, trimmed)
		return
	}

	// Normal message -> Enqueue as Spark task
	e.enqueueCommandTask(ctx, msgID, senderPhone, trimmed)
}

func (e *Engine) handleSlashCommand(ctx context.Context, senderPhone, commandText string) {
	parts := strings.Fields(commandText)
	cmd := strings.ToLower(parts[0])

	switch cmd {
	case "/help":
		_ = e.waSender.SendTextMessage(ctx, senderPhone, message.FormatHelp())

	case "/ping":
		reply := fmt.Sprintf("🏓 *Pong!*\n\nSparkWA is responsive.\nTimestamp: %s UTC", time.Now().UTC().Format(time.RFC3339))
		_ = e.waSender.SendTextMessage(ctx, senderPhone, reply)

	case "/status":
		active, _ := e.store.GetActiveTasks(ctx)
		waConn := e.waSender != nil && e.waSender.IsLoggedIn()
		gConn := false
		if e.gmailClient != nil {
			_, err := e.gmailClient.VerifyConnectivity(ctx, e.config.BridgeGmailAddress)
			gConn = (err == nil)
		}
		capReport, _ := e.capManager.Evaluate(ctx)
		capState := config.CapNotConfigured
		if capReport != nil {
			capState = capReport.State
		}

		statusMsg := message.FormatStatusReport(
			waConn,
			e.config.BotPhoneNumber,
			e.config.AuthorizedPhoneNumber,
			gConn,
			e.config.BridgeGmailAddress,
			e.config.SparkGmailAddress,
			capState,
			len(active),
		)
		_ = e.waSender.SendTextMessage(ctx, senderPhone, statusMsg)

	case "/jobs":
		recent, err := e.store.ListRecentTasks(ctx, 5)
		if err != nil {
			_ = e.waSender.SendTextMessage(ctx, senderPhone, fmt.Sprintf("❌ Error listing tasks: %v", err))
			return
		}
		_ = e.waSender.SendTextMessage(ctx, senderPhone, message.FormatJobList(recent))

	case "/job":
		if len(parts) < 2 {
			_ = e.waSender.SendTextMessage(ctx, senderPhone, "Usage: `/job <task-id>`")
			return
		}
		jobID := parts[1]
		// Match by exact ID or ID prefix
		t, err := e.findTaskByIDOrPrefix(ctx, jobID)
		if err != nil {
			_ = e.waSender.SendTextMessage(ctx, senderPhone, fmt.Sprintf("❌ Task not found: `%s`", jobID))
			return
		}
		_ = e.waSender.SendTextMessage(ctx, senderPhone, message.FormatJobDetail(t))

	case "/cancel":
		if len(parts) < 2 {
			_ = e.waSender.SendTextMessage(ctx, senderPhone, "Usage: `/cancel <task-id>`")
			return
		}
		jobID := parts[1]
		t, err := e.findTaskByIDOrPrefix(ctx, jobID)
		if err != nil {
			_ = e.waSender.SendTextMessage(ctx, senderPhone, fmt.Sprintf("❌ Task not found: `%s`", jobID))
			return
		}
		if err := e.store.CancelTask(ctx, t.ID); err != nil {
			_ = e.waSender.SendTextMessage(ctx, senderPhone, fmt.Sprintf("❌ Could not cancel `%s`: %v", t.ID, err))
			return
		}
		_ = e.waSender.SendTextMessage(ctx, senderPhone, fmt.Sprintf("🚫 Task `%s` has been cancelled.", t.ID))

	default:
		_ = e.waSender.SendTextMessage(ctx, senderPhone, fmt.Sprintf("Unknown command `%s`. Send `/help` for available commands.", cmd))
	}
}

func (e *Engine) enqueueCommandTask(ctx context.Context, waMsgID, senderPhone, commandText string) {
	// Length check
	if len(commandText) > e.config.MaxMessageLength {
		_ = e.waSender.SendTextMessage(ctx, senderPhone, fmt.Sprintf("❌ Message exceeds maximum allowed length (%d characters).", e.config.MaxMessageLength))
		return
	}

	// Concurrency check
	activeTasks, err := e.store.GetActiveTasks(ctx)
	if err == nil && len(activeTasks) >= e.config.MaxConcurrentJobs {
		_ = e.waSender.SendTextMessage(ctx, senderPhone, fmt.Sprintf("⏳ System queue is busy (%d active jobs). Please wait for ongoing tasks to finish.", len(activeTasks)))
		return
	}

	// Generate clean 8-character hex ID (e.g. 256c0c48)
	rawUUID := strings.ReplaceAll(uuid.New().String(), "-", "")
	taskID := rawUUID[:8]
	task := &store.Task{
		ID:          taskID,
		SenderPhone: senderPhone,
		WAMessageID: waMsgID,
		Command:     commandText,
		State:       store.StateQueued,
		CreatedAt:   time.Now().UTC(),
	}

	if err := e.store.CreateTask(ctx, task); err != nil {
		log.Printf("[Worker] Failed to persist task %s: %v", taskID, err)
		_ = e.waSender.SendTextMessage(ctx, senderPhone, "❌ Failed to create task in local database.")
		return
	}

	// Submit task via Gmail to Spark
	e.submitTaskToGmail(ctx, task)
}

func (e *Engine) submitTaskToGmail(ctx context.Context, task *store.Task) {
	if e.gmailClient == nil {
		_ = e.store.UpdateTaskState(ctx, task.ID, store.StateFailed, "", "Gmail client not initialized")
		_ = e.waSender.SendTextMessage(ctx, task.SenderPhone, message.FormatFailure(task.ID, "Gmail client not initialized. Run 'sparkwa gmail-auth'."))
		return
	}

	cmd := &spark.Command{
		RequestID:        task.ID,
		AuthorizedSender: task.SenderPhone,
		Timestamp:        task.CreatedAt,
		Content:          task.Command,
	}

	subject, body, err := spark.FormatCommand(cmd)
	if err != nil {
		_ = e.store.UpdateTaskState(ctx, task.ID, store.StateFailed, "", err.Error())
		_ = e.waSender.SendTextMessage(ctx, task.SenderPhone, message.FormatFailure(task.ID, err.Error()))
		return
	}

	msgID, threadID, err := e.gmailClient.SendCommand(
		ctx,
		e.config.BridgeGmailAddress,
		e.config.SparkGmailAddress,
		subject,
		body,
		task.ID,
		task.SenderPhone,
	)
	if err != nil {
		log.Printf("[Worker] Gmail send failed for task %s: %v", task.ID, err)
		_ = e.store.IncrementRetry(ctx, task.ID, err.Error())
		if task.RetryCount >= e.config.MaxTaskRetries {
			_ = e.store.UpdateTaskState(ctx, task.ID, store.StateFailed, "", fmt.Sprintf("Gmail send failed after %d retries: %v", task.RetryCount, err))
			_ = e.waSender.SendTextMessage(ctx, task.SenderPhone, message.FormatFailure(task.ID, "Failed to send command to Gemini Spark via Gmail."))
		}
		return
	}

	_ = e.store.UpdateGmailIdentifiers(ctx, task.ID, msgID, threadID)
	// Mark outbound email processed immediately so poller never ingests its own sent messages
	_ = e.store.MarkMessageProcessed(ctx, msgID, "gmail")
	_ = e.store.UpdateTaskState(ctx, task.ID, store.StateAwaitingSpark, "", "")

	log.Printf("[Worker] Forwarded task %s to Gemini Spark via Gmail (%s -> %s, MsgID: %s)", task.ID, e.config.BridgeGmailAddress, e.config.SparkGmailAddress, msgID)

	// Send submission acknowledgment back to WhatsApp Number 2
	ackMsg := message.FormatSubmissionAck(task.ID, task.Command)
	_ = e.waSender.SendTextMessage(ctx, task.SenderPhone, ackMsg)
}

func (e *Engine) pollingLoop(ctx context.Context) {
	ticker := time.NewTicker(e.config.GmailPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			e.pollForSparkReplies(ctx)
		case <-e.stopChan:
			return
		case <-ctx.Done():
			return
		}
	}
}

func (e *Engine) pollForSparkReplies(ctx context.Context) {
	if e.gmailPoller == nil {
		return
	}

	replies, err := e.gmailPoller.PollReplies(ctx)
	if err != nil {
		log.Printf("[Worker] Polling error: %v", err)
		return
	}

	for _, reply := range replies {
		e.processDiscoveredReply(ctx, reply)
	}
}

func (e *Engine) processDiscoveredReply(ctx context.Context, reply *gmail.DiscoveredReply) {
	task, err := e.store.GetTask(ctx, reply.RequestID)
	if err != nil {
		// No matching task for this request ID
		return
	}

	// Only process replies for tasks that are currently waiting
	if task.State != store.StateAwaitingSpark && task.State != store.StateSubmitted {
		return
	}

	parsed, err := spark.ParseResponse(task.ID, reply.BodyText)
	if err != nil {
		log.Printf("[Worker] Error parsing response for task %s: %v", task.ID, err)
		return
	}

	// Deduplicate reply
	_ = e.store.MarkMessageProcessed(ctx, reply.GmailMessageID, "gmail")

	switch parsed.Status {
	case spark.StatusApprovalRequired:
		_ = e.store.UpdateTaskState(ctx, task.ID, store.StateApprovalRequired, parsed.ResultText, "")
		promptMsg := message.FormatApprovalNeeded(task.ID, parsed.ApprovalPrompt)
		_ = e.waSender.SendTextMessage(ctx, task.SenderPhone, promptMsg)

	case spark.StatusSuccess:
		_ = e.store.UpdateTaskState(ctx, task.ID, store.StateCompleted, parsed.ResultText, "")
		resultMsg := message.FormatCompletion(task.ID, parsed.ResultText)
		_ = e.waSender.SendTextMessage(ctx, task.SenderPhone, resultMsg)

		// Record that return path has been verified!
		_ = e.capManager.SetState(ctx, config.CapResponsePathVerified)

	case spark.StatusFailed:
		_ = e.store.UpdateTaskState(ctx, task.ID, store.StateFailed, "", parsed.ResultText)
		failMsg := message.FormatFailure(task.ID, parsed.ResultText)
		_ = e.waSender.SendTextMessage(ctx, task.SenderPhone, failMsg)
	}
}

func (e *Engine) sweeperLoop(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			expiredCount, err := e.store.ExpireOldTasks(ctx, e.config.TaskTimeout)
			if err != nil {
				log.Printf("[Worker] Sweeper error: %v", err)
			} else if expiredCount > 0 {
				log.Printf("[Worker] Sweeper marked %d tasks as EXPIRED", expiredCount)
			}
		case <-e.stopChan:
			return
		case <-ctx.Done():
			return
		}
	}
}

func (e *Engine) recoverActiveTasks(ctx context.Context) {
	active, err := e.store.GetActiveTasks(ctx)
	if err != nil {
		log.Printf("[Worker] Failed to query active tasks on startup: %v", err)
		return
	}

	if len(active) > 0 {
		log.Printf("[Worker] Recovered %d active tasks from persistent SQLite database", len(active))
		for _, t := range active {
			if t.State == store.StateQueued {
				// Re-submit queued tasks that never sent
				go e.submitTaskToGmail(ctx, t)
			}
		}
	}
}

func (e *Engine) findTaskByIDOrPrefix(ctx context.Context, queryID string) (*store.Task, error) {
	// Try direct lookup
	t, err := e.store.GetTask(ctx, queryID)
	if err == nil {
		return t, nil
	}

	// Try prefix match from recent tasks
	recent, err := e.store.ListRecentTasks(ctx, 50)
	if err != nil {
		return nil, err
	}

	for _, task := range recent {
		if strings.HasPrefix(task.ID, queryID) {
			return task, nil
		}
	}

	return nil, store.ErrTaskNotFound
}
