package whatsapp

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/mdp/qrterminal/v3"
	_ "modernc.org/sqlite"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"

	"github.com/sparkwa/sparkwa-lite/internal/identity"
	"github.com/sparkwa/sparkwa-lite/internal/message"
)

// MessageHandler is the callback function invoked when a verified command from Number 2 arrives.
type MessageHandler func(ctx context.Context, msgID string, senderPhone string, text string)

// Client wraps whatsmeow with persistent SQLite storage and terminal QR login.
type Client struct {
	client    *whatsmeow.Client
	container *sqlstore.Container
	verifier  *identity.Verifier
	handler   MessageHandler
	logger    waLog.Logger
	mu        sync.Mutex
	running   bool
}

// NewClient initializes the whatsmeow client using SQLite session persistence.
func NewClient(ctx context.Context, sessionDBPath string, verifier *identity.Verifier, logLevel string) (*Client, error) {
	if !filepath.IsAbs(sessionDBPath) && sessionDBPath != ":memory:" {
		abs, err := filepath.Abs(sessionDBPath)
		if err == nil {
			sessionDBPath = abs
		}
	}

	dir := filepath.Dir(sessionDBPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	var waLogLevel string
	switch logLevel {
	case "debug":
		waLogLevel = "DEBUG"
	case "warn":
		waLogLevel = "WARN"
	case "error":
		waLogLevel = "ERROR"
	default:
		waLogLevel = "INFO"
	}

	logger := waLog.Stdout("WhatsMeow", waLogLevel, true)
	dsn := fmt.Sprintf("file:%s?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(ON)", sessionDBPath)

	container, err := sqlstore.New(ctx, "sqlite", dsn, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to open whatsmeow session store: %w", err)
	}

	deviceStore, err := container.GetFirstDevice(ctx)
	if err != nil {
		_ = container.Close()
		return nil, fmt.Errorf("failed to retrieve device store: %w", err)
	}

	waClient := whatsmeow.NewClient(deviceStore, logger)

	c := &Client{
		client:    waClient,
		container: container,
		verifier:  verifier,
		logger:    logger,
	}

	waClient.AddEventHandler(c.handleEvent)

	return c, nil
}

// SetMessageHandler registers the callback for verified user messages.
func (c *Client) SetMessageHandler(h MessageHandler) {
	c.handler = h
}

// IsLoggedIn checks whether a valid WhatsApp session exists.
func (c *Client) IsLoggedIn() bool {
	return c.client.Store.ID != nil
}

// DeviceJID returns the bot's paired JID or empty string.
func (c *Client) DeviceJID() string {
	if c.client.Store.ID == nil {
		return ""
	}
	return c.client.Store.ID.String()
}

// LoginQR presents a terminal QR code if not logged in, or connects immediately.
func (c *Client) LoginQR(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.client.Store.ID != nil {
		fmt.Printf("✓ Already paired to WhatsApp device: %s\n", c.client.Store.ID.String())
		return c.client.Connect()
	}

	qrChan, err := c.client.GetQRChannel(ctx)
	if err != nil {
		return fmt.Errorf("failed to get QR channel: %w", err)
	}

	if err := c.client.Connect(); err != nil {
		return fmt.Errorf("failed to connect to WhatsApp servers: %w", err)
	}

	fmt.Println("================================================================================")
	fmt.Println("                WHATSAPP LINKED DEVICES TERMINAL PAIRING")
	fmt.Println("================================================================================")
	fmt.Println("1. Open WhatsApp on phone Number 1 (the bot account).")
	fmt.Println("2. Go to Settings > Linked Devices > Link a Device.")
	fmt.Println("3. Scan the QR code displayed below in your SSH terminal:")
	fmt.Println("================================================================================")

	for evt := range qrChan {
		switch evt.Event {
		case "code":
			fmt.Println("\nScan this QR code:")
			qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
		case "success":
			fmt.Println("\n✓ WhatsApp device pairing successful!")
			return nil
		case "timeout":
			return errors.New("pairing QR code timed out. Re-run 'sparkwa login'")
		default:
			fmt.Printf("Login event: %s\n", evt.Event)
		}
	}

	return nil
}

// Connect connects to the WhatsApp network.
func (c *Client) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.running {
		return nil
	}

	if err := c.client.Connect(); err != nil {
		return err
	}

	c.running = true
	return nil
}

// Disconnect gracefully disconnects from WhatsApp.
func (c *Client) Disconnect() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.running {
		return
	}

	c.client.Disconnect()
	c.running = false
}

// Close closes the underlying SQLite database container.
func (c *Client) Close() error {
	c.Disconnect()
	return c.container.Close()
}

// SendTextMessage sends a text message to the authorized user's phone, automatically chunking long replies.
func (c *Client) SendTextMessage(ctx context.Context, recipientPhone, text string) error {
	norm, err := identity.NormalizeE164(recipientPhone)
	if err != nil {
		return fmt.Errorf("invalid recipient phone: %w", err)
	}

	// Build recipient JID (+1234567890 -> 1234567890@s.whatsapp.net)
	userDigits := strings.TrimPrefix(norm, "+")
	recipientJID := types.NewJID(userDigits, types.DefaultUserServer)

	chunks := message.SplitText(text, message.DefaultMaxChunkSize)

	for i, chunk := range chunks {
		msg := &waE2E.Message{
			Conversation: proto.String(chunk),
		}

		_, err := c.client.SendMessage(ctx, recipientJID, msg)
		if err != nil {
			return fmt.Errorf("failed to send message chunk %d/%d: %w", i+1, len(chunks), err)
		}

		// Small delay between chunks to prevent message inversion on WhatsApp network
		if len(chunks) > 1 && i < len(chunks)-1 {
			select {
			case <-time.After(200 * time.Millisecond):
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	return nil
}

// handleEvent handles whatsmeow incoming events.
func (c *Client) handleEvent(evt any) {
	switch v := evt.(type) {
	case *events.Message:
		c.handleIncomingMessage(v)
	case *events.LoggedOut:
		c.logger.Errorf("WhatsApp session logged out or revoked: %s", v.Reason)
	case *events.Connected:
		c.logger.Infof("Connected to WhatsApp multi-device gateway.")
	case *events.Disconnected:
		c.logger.Warnf("Disconnected from WhatsApp. Automatic reconnection in progress...")
	}
}

// handleIncomingMessage applies security filtering and passes authorized messages to the handler.
func (c *Client) handleIncomingMessage(evt *events.Message) {
	ctx := context.Background()

	// 1. Filter groups, status broadcasts, self messages
	isStatus := evt.Info.Chat == types.StatusBroadcastJID ||
		evt.Info.Chat.Server == types.BroadcastServer ||
		evt.Info.IsIncomingBroadcast()

	if err := c.verifier.CheckMessageFilter(evt.Info.IsGroup, isStatus, evt.Info.IsFromMe); err != nil {
		// Silently ignore per requirements
		return
	}

	// 2. Authorize sender (strictly Number 2 only)
	senderPhone, err := c.verifier.Authorize(ctx, evt.Info.Sender)
	if err != nil {
		// Silently ignore unauthorized senders
		return
	}

	// 3. Extract text content
	text := extractTextFromMessage(evt.Message)
	if strings.TrimSpace(text) == "" {
		// Version 1 only processes text messages
		return
	}

	// 4. Pass to worker handler
	if c.handler != nil {
		c.handler(ctx, evt.Info.ID, senderPhone, strings.TrimSpace(text))
	}
}

// extractTextFromMessage extracts conversation or extended text string.
func extractTextFromMessage(msg *waE2E.Message) string {
	if msg == nil {
		return ""
	}
	if msg.Conversation != nil && *msg.Conversation != "" {
		return *msg.Conversation
	}
	if msg.ExtendedTextMessage != nil && msg.ExtendedTextMessage.Text != nil {
		return *msg.ExtendedTextMessage.Text
	}
	return ""
}
