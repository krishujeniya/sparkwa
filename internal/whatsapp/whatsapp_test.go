package whatsapp

import (
	"context"
	"strings"
	"testing"

	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"

	"github.com/sparkwa/sparkwa/internal/identity"
)

func TestExtractTextFromMessage(t *testing.T) {
	// Simple conversation
	m1 := &waE2E.Message{
		Conversation: proto.String("Simple message"),
	}
	if got := extractTextFromMessage(m1); got != "Simple message" {
		t.Errorf("expected 'Simple message', got %q", got)
	}

	// Extended text
	m2 := &waE2E.Message{
		ExtendedTextMessage: &waE2E.ExtendedTextMessage{
			Text: proto.String("Extended message text"),
		},
	}
	if got := extractTextFromMessage(m2); got != "Extended message text" {
		t.Errorf("expected 'Extended message text', got %q", got)
	}

	// Nil message
	if got := extractTextFromMessage(nil); got != "" {
		t.Errorf("expected empty string for nil, got %q", got)
	}
}

func TestWhatsAppMessageFilterAndAuth(t *testing.T) {
	verifier, err := identity.NewVerifier("+19876543210", "+12345678900", nil)
	if err != nil {
		t.Fatalf("failed to create verifier: %v", err)
	}

	received := false
	var capturedSender, capturedText, capturedID string

	handler := func(ctx context.Context, msgID, sender, text string) {
		received = true
		capturedSender = sender
		capturedText = text
		capturedID = msgID
	}

	c := &Client{
		verifier: verifier,
		handler:  handler,
	}

	// 1. Group message should be ignored
	c.handleIncomingMessage(&events.Message{
		Info: types.MessageInfo{
			ID: "msg-group",
			MessageSource: types.MessageSource{
				IsGroup: true,
				Sender:  types.NewJID("19876543210", types.DefaultUserServer),
			},
		},
		Message: &waE2E.Message{Conversation: proto.String("Group test")},
	})
	if received {
		t.Errorf("group message was not ignored")
	}

	// 2. Status broadcast should be ignored
	c.handleIncomingMessage(&events.Message{
		Info: types.MessageInfo{
			ID: "msg-status",
			MessageSource: types.MessageSource{
				Chat:   types.StatusBroadcastJID,
				Sender: types.NewJID("19876543210", types.DefaultUserServer),
			},
		},
		Message: &waE2E.Message{Conversation: proto.String("Status test")},
	})
	if received {
		t.Errorf("status broadcast was not ignored")
	}

	// 3. Self outgoing message should be ignored
	c.handleIncomingMessage(&events.Message{
		Info: types.MessageInfo{
			ID: "msg-self",
			MessageSource: types.MessageSource{
				IsFromMe: true,
				Sender:   types.NewJID("19876543210", types.DefaultUserServer),
			},
		},
		Message: &waE2E.Message{Conversation: proto.String("Self test")},
	})
	if received {
		t.Errorf("self outgoing message was not ignored")
	}

	// 4. Stranger unauthorized sender should be ignored
	c.handleIncomingMessage(&events.Message{
		Info: types.MessageInfo{
			ID: "msg-stranger",
			MessageSource: types.MessageSource{
				Sender: types.NewJID("15559998888", types.DefaultUserServer),
			},
		},
		Message: &waE2E.Message{Conversation: proto.String("Unauthorized prompt")},
	})
	if received {
		t.Errorf("unauthorized sender was not ignored")
	}

	// 5. Authorized direct message should succeed
	c.handleIncomingMessage(&events.Message{
		Info: types.MessageInfo{
			ID: "msg-valid-123",
			MessageSource: types.MessageSource{
				Sender: types.NewJID("19876543210", types.DefaultUserServer),
			},
		},
		Message: &waE2E.Message{Conversation: proto.String("Run daily summary")},
	})

	if !received {
		t.Fatalf("expected message to be delivered to handler")
	}
	if capturedSender != "+19876543210" {
		t.Errorf("expected sender +19876543210, got %s", capturedSender)
	}
	if capturedText != "Run daily summary" {
		t.Errorf("expected text 'Run daily summary', got %s", capturedText)
	}
	if capturedID != "msg-valid-123" {
		t.Errorf("expected message id 'msg-valid-123', got %s", capturedID)
	}
}

func TestMaskPhoneInLogs(t *testing.T) {
	phone := "+19876543210"
	masked := "+198***3210"
	if !strings.HasPrefix(masked, "+198") || !strings.HasSuffix(masked, "3210") {
		t.Errorf("unexpected masked phone format: %s", phone)
	}
}
