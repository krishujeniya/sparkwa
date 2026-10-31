package spark

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

var (
	ErrInvalidRequestID = errors.New("invalid or missing request ID")
	ErrEmptyCommand     = errors.New("command cannot be empty")
)

// Command represents an outgoing prompt / task dispatched to Gemini Spark via Gmail.
type Command struct {
	RequestID        string
	AuthorizedSender string
	Timestamp        time.Time
	Content          string
}

// FormatCommand serializes the command into the standard SparkWA envelope.
func FormatCommand(cmd *Command) (subject string, body string, err error) {
	if strings.TrimSpace(cmd.RequestID) == "" {
		return "", "", ErrInvalidRequestID
	}
	if strings.TrimSpace(cmd.Content) == "" {
		return "", "", ErrEmptyCommand
	}
	if cmd.Timestamp.IsZero() {
		cmd.Timestamp = time.Now().UTC()
	}

	// Short subject summary for email readability and Spark Gmail filter matching
	shortSummary := strings.ReplaceAll(cmd.Content, "\n", " ")
	if len(shortSummary) > 60 {
		shortSummary = shortSummary[:57] + "..."
	}
	subject = fmt.Sprintf("[SPARKWA:%s] %s", cmd.RequestID, shortSummary)

	body = fmt.Sprintf(
		"BEGIN SPARKWA COMMAND\n"+
			"Request ID: %s\n"+
			"Authorized sender: %s\n"+
			"Timestamp: %s\n"+
			"Command: %s\n"+
			"END SPARKWA COMMAND\n",
		cmd.RequestID,
		cmd.AuthorizedSender,
		cmd.Timestamp.Format(time.RFC3339),
		strings.TrimSpace(cmd.Content),
	)

	return subject, body, nil
}

// ResponseStatus indicates the result of Spark's execution.
type ResponseStatus string

const (
	StatusSuccess          ResponseStatus = "SUCCESS"
	StatusApprovalRequired ResponseStatus = "APPROVAL_REQUIRED"
	StatusFailed           ResponseStatus = "FAILED"
)

// Response represents a parsed reply received from Spark via email or draft.
type Response struct {
	RequestID        string
	Status           ResponseStatus
	ResultText       string
	ApprovalPrompt   string
	IsStructured     bool
	SourceMessageID  string
	SourceThreadID   string
}

var (
	reqIDRegex     = regexp.MustCompile(`\[SPARKWA:([a-zA-Z0-9_-]+)\]`)
	blockReqID     = regexp.MustCompile(`(?i)Request ID:\s*([a-zA-Z0-9_-]+)`)
	blockStatus    = regexp.MustCompile(`(?i)Status:\s*(SUCCESS|APPROVAL_REQUIRED|FAILED)`)
	approvalPhrases = []string{
		"do you want me to",
		"please confirm if",
		"approval required",
		"requires your confirmation",
		"would you like me to proceed",
		"reply yes to confirm",
		"awaiting your approval",
	}
)

// ExtractRequestIDFromSubject extracts the correlation ID from an email subject line.
func ExtractRequestIDFromSubject(subject string) string {
	matches := reqIDRegex.FindStringSubmatch(subject)
	if len(matches) == 2 {
		return matches[1]
	}
	return ""
}

// ParseResponse parses an email body or draft content to match it against an expected request.
func ParseResponse(expectedReqID, bodyText string) (*Response, error) {
	cleanBody := strings.TrimSpace(bodyText)
	if cleanBody == "" {
		return nil, errors.New("empty response body")
	}

	// 1. Check for formal BEGIN SPARKWA RESPONSE ... END SPARKWA RESPONSE block
	startIdx := strings.Index(cleanBody, "BEGIN SPARKWA RESPONSE")
	endIdx := strings.Index(cleanBody, "END SPARKWA RESPONSE")

	if startIdx != -1 && endIdx != -1 && endIdx > startIdx {
		block := cleanBody[startIdx+len("BEGIN SPARKWA RESPONSE") : endIdx]
		reqMatch := blockReqID.FindStringSubmatch(block)
		if len(reqMatch) == 2 && (expectedReqID == "" || reqMatch[1] == expectedReqID) {
			res := &Response{
				RequestID:    reqMatch[1],
				Status:       StatusSuccess,
				IsStructured: true,
			}

			statusMatch := blockStatus.FindStringSubmatch(block)
			if len(statusMatch) == 2 {
				res.Status = ResponseStatus(strings.ToUpper(statusMatch[1]))
			}

			// Extract result section
			resIdx := strings.Index(block, "Result:")
			if resIdx != -1 {
				res.ResultText = strings.TrimSpace(block[resIdx+len("Result:"):])
			} else {
				res.ResultText = strings.TrimSpace(block)
			}

			// If status is approval required
			if res.Status == StatusApprovalRequired {
				res.ApprovalPrompt = res.ResultText
			}

			return res, nil
		}
	}

	// 2. Fallback: Spark replied as a standard email in the thread
	// Verify whether the body contains the request ID or if caller validated thread
	extractedID := ""
	reqMatch := blockReqID.FindStringSubmatch(cleanBody)
	if len(reqMatch) == 2 {
		extractedID = reqMatch[1]
	} else {
		extractedID = expectedReqID
	}

	if expectedReqID != "" && extractedID != expectedReqID {
		return nil, fmt.Errorf("response request ID %q does not match expected %q", extractedID, expectedReqID)
	}

	res := &Response{
		RequestID:    extractedID,
		Status:       StatusSuccess,
		ResultText:   cleanBody,
		IsStructured: false,
	}

	// Detect whether Spark is asking for approval before executing an action
	lower := strings.ToLower(cleanBody)
	for _, phrase := range approvalPhrases {
		if strings.Contains(lower, phrase) {
			res.Status = StatusApprovalRequired
			res.ApprovalPrompt = cleanBody
			break
		}
	}

	return res, nil
}
