package spark

import (
	"strings"
	"testing"
	"time"
)

func TestFormatCommand(t *testing.T) {
	cmd := &Command{
		RequestID:        "test-req-123",
		AuthorizedSender: "+19876543210",
		Timestamp:        time.Date(2026, 10, 31, 12, 0, 0, 0, time.UTC),
		Content:          "Please draft a summary of Q3 client invoices.",
	}

	subject, body, err := FormatCommand(cmd)
	if err != nil {
		t.Fatalf("FormatCommand failed: %v", err)
	}

	if !strings.HasPrefix(subject, "[SPARKWA:test-req-123]") {
		t.Errorf("subject missing request ID prefix: %s", subject)
	}

	if !strings.Contains(body, "BEGIN SPARKWA COMMAND") || !strings.Contains(body, "END SPARKWA COMMAND") {
		t.Errorf("body missing envelope markers: %s", body)
	}

	if !strings.Contains(body, "Request ID: test-req-123") {
		t.Errorf("body missing request ID: %s", body)
	}

	if !strings.Contains(body, "Authorized sender: +1987") {
		t.Errorf("body missing authorized sender: %s", body)
	}

	if !strings.Contains(body, "Please draft a summary of Q3 client invoices.") {
		t.Errorf("body missing command content: %s", body)
	}
}

func TestExtractRequestIDFromSubject(t *testing.T) {
	tests := []struct {
		subject string
		want    string
	}{
		{"Re: [SPARKWA:uuid-9876-abcd] Command execution", "uuid-9876-abcd"},
		{"[SPARKWA:simple-id] test", "simple-id"},
		{"Fwd: [SPARKWA:req-1] Spark response", "req-1"},
		{"Ordinary email without tag", ""},
		{"[OTHER:123] not matching", ""},
	}

	for _, tt := range tests {
		got := ExtractRequestIDFromSubject(tt.subject)
		if got != tt.want {
			t.Errorf("ExtractRequestIDFromSubject(%q) = %q, want %q", tt.subject, got, tt.want)
		}
	}
}

func TestParseResponseStructured(t *testing.T) {
	body := `
Hello,

BEGIN SPARKWA RESPONSE
Request ID: req-456
Status: SUCCESS
Result: All 12 invoices have been compiled into your Google Sheet "Q3 Billing".
END SPARKWA RESPONSE

Thanks,
Gemini Spark
`

	resp, err := ParseResponse("req-456", body)
	if err != nil {
		t.Fatalf("ParseResponse failed: %v", err)
	}

	if !resp.IsStructured {
		t.Errorf("expected IsStructured to be true")
	}
	if resp.RequestID != "req-456" {
		t.Errorf("got reqID %q, want req-456", resp.RequestID)
	}
	if resp.Status != StatusSuccess {
		t.Errorf("got status %v, want StatusSuccess", resp.Status)
	}
	if resp.ResultText != "All 12 invoices have been compiled into your Google Sheet \"Q3 Billing\"." {
		t.Errorf("unexpected result text: %q", resp.ResultText)
	}
}

func TestParseResponseApprovalRequired(t *testing.T) {
	// 1. Structured approval required
	structuredBody := `
BEGIN SPARKWA RESPONSE
Request ID: req-789
Status: APPROVAL_REQUIRED
Result: I am ready to send the contract to client@example.com. Do you want me to proceed?
END SPARKWA RESPONSE
`
	resp, err := ParseResponse("req-789", structuredBody)
	if err != nil {
		t.Fatalf("ParseResponse structured approval failed: %v", err)
	}
	if resp.Status != StatusApprovalRequired {
		t.Errorf("expected StatusApprovalRequired, got %s", resp.Status)
	}

	// 2. Unstructured fallback approval required
	freeformBody := `
I reviewed your files. I found 3 sensitive documents.
Would you like me to proceed with archiving them?
Request ID: req-freeform
`
	respFreeform, err := ParseResponse("req-freeform", freeformBody)
	if err != nil {
		t.Fatalf("ParseResponse freeform failed: %v", err)
	}
	if respFreeform.Status != StatusApprovalRequired {
		t.Errorf("expected StatusApprovalRequired for freeform text with approval phrase, got %s", respFreeform.Status)
	}
}

func TestParseResponseRejectsCommandEcho(t *testing.T) {
	cmdEcho := `
BEGIN SPARKWA COMMAND
Request ID: 256c0c48-test
Authorized sender: +1987••••••10
Timestamp: 2026-11-03T04:01:36Z
Command: Hi
END SPARKWA COMMAND
`
	_, err := ParseResponse("256c0c48-test", cmdEcho)
	if err == nil {
		t.Fatalf("expected ParseResponse to reject pure command echo, but it accepted it as success")
	}
}

func TestParseResponseCleansQuotedEmail(t *testing.T) {
	emailBody := `Hello! How can I help you today?

On Tue, Nov 3, 2026 at 9:31 AM bridge@example.com wrote:
> BEGIN SPARKWA COMMAND
> Request ID: 256c0c48
> Authorized sender: +1987••••••10
> Timestamp: 2026-11-03T04:01:36Z
> Command: Hi
> END SPARKWA COMMAND
`
	resp, err := ParseResponse("256c0c48", emailBody)
	if err != nil {
		t.Fatalf("ParseResponse failed: %v", err)
	}
	if resp.ResultText != "Hello! How can I help you today?" {
		t.Errorf("expected clean result, got: %q", resp.ResultText)
	}
}

func TestParseResponseMismatchedID(t *testing.T) {
	body := `
Request ID: other-id
Result: Some data
`
	_, err := ParseResponse("expected-id", body)
	if err == nil {
		t.Errorf("expected error for mismatched request ID, got nil")
	}
}
