package gmail

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/sparkwa/sparkwa/internal/spark"
	"github.com/sparkwa/sparkwa/internal/store"
	"google.golang.org/api/gmail/v1"
)

// DiscoveredReply contains the raw text and metadata of a message found in Gmail.
type DiscoveredReply struct {
	RequestID      string
	GmailMessageID string
	GmailThreadID  string
	From           string
	Subject        string
	BodyText       string
	IsDraft        bool
}

// Poller searches for replies sent from the Spark agent.
type Poller struct {
	client            *Client
	store             *store.Store
	sparkEmail        string
	bridgeEmail       string
	sparkResponseMode string
}

// NewPoller initializes a new Gmail Poller.
func NewPoller(client *Client, store *store.Store, sparkEmail, bridgeEmail, responseMode string) *Poller {
	return &Poller{
		client:            client,
		store:             store,
		sparkEmail:        strings.ToLower(strings.TrimSpace(sparkEmail)),
		bridgeEmail:       strings.ToLower(strings.TrimSpace(bridgeEmail)),
		sparkResponseMode: responseMode,
	}
}

// PollReplies checks both the inbox for incoming messages from Spark, and optionally drafts.
func (p *Poller) PollReplies(ctx context.Context) ([]*DiscoveredReply, error) {
	var results []*DiscoveredReply

	// 1. Search messages in Gmail
	if p.sparkResponseMode != "draft_only" {
		msgReplies, err := p.pollMessages(ctx)
		if err != nil {
			return nil, fmt.Errorf("polling messages failed: %w", err)
		}
		results = append(results, msgReplies...)
	}

	// 2. Search drafts if enabled
	if p.sparkResponseMode == "draft_or_email" || p.sparkResponseMode == "draft_only" {
		draftReplies, err := p.pollDrafts(ctx)
		if err != nil {
			// Do not fail entire poll if drafts scope is not granted or draft check fails
			_ = err
		} else {
			results = append(results, draftReplies...)
		}
	}

	return results, nil
}

func containsLabel(labels []string, target string) bool {
	for _, l := range labels {
		if strings.EqualFold(l, target) {
			return true
		}
	}
	return false
}

func (p *Poller) pollMessages(ctx context.Context) ([]*DiscoveredReply, error) {
	// Query messages matching SPARKWA subject tag, filtering out sent emails
	var query string
	if p.sparkEmail != "" && !strings.EqualFold(p.sparkEmail, p.bridgeEmail) {
		query = fmt.Sprintf("from:%s subject:SPARKWA", p.sparkEmail)
	} else {
		query = "-from:me subject:SPARKWA"
	}
	listCall := p.client.service.Users.Messages.List("me").Q(query).MaxResults(20)

	resp, err := listCall.Context(ctx).Do()
	if err != nil {
		return nil, err
	}

	var replies []*DiscoveredReply

	for _, msgSummary := range resp.Messages {
		// Check deduplication
		alreadyProcessed, err := p.store.IsMessageProcessed(ctx, msgSummary.Id, "gmail")
		if err != nil || alreadyProcessed {
			continue
		}

		fullMsg, err := p.client.service.Users.Messages.Get("me", msgSummary.Id).Format("full").Context(ctx).Do()
		if err != nil {
			continue
		}

		// Never ingest outbound sent messages as replies
		if containsLabel(fullMsg.LabelIds, "SENT") && !containsLabel(fullMsg.LabelIds, "INBOX") {
			continue
		}

		reply, err := p.extractReplyFromMessage(fullMsg, false)
		if err != nil {
			continue
		}

		// Security: Verify that the email was actually sent by the Spark Gmail account
		fromLower := strings.ToLower(reply.From)
		if !strings.EqualFold(p.sparkEmail, p.bridgeEmail) {
			if !strings.Contains(fromLower, p.sparkEmail) {
				// Untrusted sender: ignore message to prevent spoofed command injection
				continue
			}
		} else {
			if containsLabel(fullMsg.LabelIds, "SENT") {
				continue
			}
		}

		replies = append(replies, reply)
	}

	return replies, nil
}

func (p *Poller) pollDrafts(ctx context.Context) ([]*DiscoveredReply, error) {
	draftList, err := p.client.service.Users.Drafts.List("me").MaxResults(20).Context(ctx).Do()
	if err != nil {
		return nil, err
	}

	var replies []*DiscoveredReply

	for _, d := range draftList.Drafts {
		if d.Message == nil {
			continue
		}

		alreadyProcessed, err := p.store.IsMessageProcessed(ctx, d.Message.Id, "gmail_draft")
		if err != nil || alreadyProcessed {
			continue
		}

		fullDraft, err := p.client.service.Users.Drafts.Get("me", d.Id).Format("full").Context(ctx).Do()
		if err != nil || fullDraft.Message == nil {
			continue
		}

		reply, err := p.extractReplyFromMessage(fullDraft.Message, true)
		if err != nil {
			continue
		}

		replies = append(replies, reply)
	}

	return replies, nil
}

func (p *Poller) extractReplyFromMessage(msg *gmail.Message, isDraft bool) (*DiscoveredReply, error) {
	var subject, from, reqIDHeader string

	if msg.Payload != nil {
		for _, h := range msg.Payload.Headers {
			switch strings.ToLower(h.Name) {
			case "subject":
				subject = h.Value
			case "from":
				from = h.Value
			case "x-sparkwa-request-id":
				reqIDHeader = h.Value
			}
		}
	}

	// Extract Request ID from header or subject line
	reqID := reqIDHeader
	if reqID == "" {
		reqID = spark.ExtractRequestIDFromSubject(subject)
	}

	if reqID == "" {
		return nil, errors.New("no SparkWA request ID found in message")
	}

	bodyText := ExtractBodyText(msg.Payload)
	if strings.TrimSpace(bodyText) == "" {
		return nil, errors.New("empty message body text")
	}

	return &DiscoveredReply{
		RequestID:      reqID,
		GmailMessageID: msg.Id,
		GmailThreadID:  msg.ThreadId,
		From:           from,
		Subject:        subject,
		BodyText:       bodyText,
		IsDraft:        isDraft,
	}, nil
}

// ExtractBodyText recursively extracts plain text from MIME payload parts.
func ExtractBodyText(payload *gmail.MessagePart) string {
	if payload == nil {
		return ""
	}

	// If single part with data
	if payload.Body != nil && payload.Body.Data != "" {
		decoded, err := base64.URLEncoding.DecodeString(payload.Body.Data)
		if err == nil {
			text := string(decoded)
			if payload.MimeType == "text/html" {
				return StripHTML(text)
			}
			return text
		}
	}

	// Prefer text/plain if multiple parts exist
	var plainText, htmlText strings.Builder
	for _, part := range payload.Parts {
		if part.MimeType == "text/plain" && part.Body != nil && part.Body.Data != "" {
			if decoded, err := base64.URLEncoding.DecodeString(part.Body.Data); err == nil {
				plainText.WriteString(string(decoded))
			}
		} else if part.MimeType == "text/html" && part.Body != nil && part.Body.Data != "" {
			if decoded, err := base64.URLEncoding.DecodeString(part.Body.Data); err == nil {
				htmlText.WriteString(StripHTML(string(decoded)))
			}
		} else if len(part.Parts) > 0 {
			nested := ExtractBodyText(part)
			plainText.WriteString(nested)
		}
	}

	if plainText.Len() > 0 {
		return plainText.String()
	}
	return htmlText.String()
}

var (
	tagRegex    = regexp.MustCompile(`<[^>]*>`)
	spaceRegex  = regexp.MustCompile(`\s+`)
	entityRegex = regexp.MustCompile(`&nbsp;|&lt;|&gt;|&amp;`)
)

// StripHTML strips HTML tags and normalizes whitespace.
func StripHTML(input string) string {
	noTags := tagRegex.ReplaceAllString(input, " ")
	replacer := strings.NewReplacer(
		"&nbsp;", " ",
		"&lt;", "<",
		"&gt;", ">",
		"&amp;", "&",
		"&quot;", "\"",
		"&#39;", "'",
	)
	decoded := replacer.Replace(noTags)
	return strings.TrimSpace(decoded)
}
