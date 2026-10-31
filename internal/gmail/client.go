package gmail

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/mail"
	"strings"

	"golang.org/x/oauth2"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

// Client encapsulates interactions with Google's Gmail API.
type Client struct {
	service *gmail.Service
}

// NewClient initializes the Gmail client using OAuth credentials and stored token.
func NewClient(ctx context.Context, oauthCfg *oauth2.Config, tokenPath string) (*Client, error) {
	tokenSource, err := GetTokenSource(ctx, oauthCfg, tokenPath)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain token source: %w", err)
	}

	srv, err := gmail.NewService(ctx, option.WithTokenSource(tokenSource))
	if err != nil {
		return nil, fmt.Errorf("failed to create gmail service: %w", err)
	}

	return &Client{service: srv}, nil
}

// Service returns the underlying Gmail API service.
func (c *Client) Service() *gmail.Service {
	return c.service
}

// VerifyConnectivity checks whether the client can communicate with the Gmail API.
func (c *Client) VerifyConnectivity(ctx context.Context, expectedEmail string) (string, error) {
	profile, err := c.service.Users.GetProfile("me").Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("gmail profile check failed: %w", err)
	}

	if expectedEmail != "" {
		if !strings.EqualFold(profile.EmailAddress, expectedEmail) {
			return profile.EmailAddress, fmt.Errorf("authenticated account (%s) does not match configured bridge address (%s)", profile.EmailAddress, expectedEmail)
		}
	}

	return profile.EmailAddress, nil
}

// BuildRFC2822Message constructs a standard MIME email format.
func BuildRFC2822Message(from, to, subject, body, requestID, senderPhone string) ([]byte, error) {
	if _, err := mail.ParseAddress(from); err != nil {
		return nil, fmt.Errorf("invalid 'from' address: %w", err)
	}
	if _, err := mail.ParseAddress(to); err != nil {
		return nil, fmt.Errorf("invalid 'to' address: %w", err)
	}

	encodedBody := base64.StdEncoding.EncodeToString([]byte(body))

	// Format MIME headers
	mimeHeader := fmt.Sprintf(
		"From: %s\r\n"+
			"To: %s\r\n"+
			"Subject: %s\r\n"+
			"X-SparkWA-Request-ID: %s\r\n"+
			"X-SparkWA-Sender: %s\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: text/plain; charset=\"UTF-8\"\r\n"+
			"Content-Transfer-Encoding: base64\r\n\r\n"+
			"%s\r\n",
		from,
		to,
		subject,
		requestID,
		senderPhone,
		encodedBody,
	)

	return []byte(mimeHeader), nil
}

// SendCommand dispatches a command email to the Spark mailbox.
func (c *Client) SendCommand(ctx context.Context, from, to, subject, body, requestID, senderPhone string) (messageID string, threadID string, err error) {
	rawMIME, err := BuildRFC2822Message(from, to, subject, body, requestID, senderPhone)
	if err != nil {
		return "", "", fmt.Errorf("failed to build MIME message: %w", err)
	}

	// Gmail API expects raw message in URL-safe base64 string
	gMsg := &gmail.Message{
		Raw: base64.URLEncoding.EncodeToString(rawMIME),
	}

	sent, err := c.service.Users.Messages.Send("me", gMsg).Context(ctx).Do()
	if err != nil {
		return "", "", fmt.Errorf("failed to send message via Gmail API: %w", err)
	}

	return sent.Id, sent.ThreadId, nil
}
