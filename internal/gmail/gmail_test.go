package gmail

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"golang.org/x/oauth2"
	"google.golang.org/api/gmail/v1"
)

func TestBuildRFC2822Message(t *testing.T) {
	from := "bridge@example.com"
	to := "spark@example.com"
	subject := "[SPARKWA:uuid-123] Test prompt"
	body := "BEGIN SPARKWA COMMAND\nRequest ID: uuid-123\nCommand: Hello Spark\nEND SPARKWA COMMAND"
	reqID := "uuid-123"
	senderPhone := "+19876543210"

	raw, err := BuildRFC2822Message(from, to, subject, body, reqID, senderPhone)
	if err != nil {
		t.Fatalf("BuildRFC2822Message failed: %v", err)
	}

	content := string(raw)
	if !strings.Contains(content, "From: bridge@example.com") {
		t.Errorf("missing From header in MIME: %s", content)
	}
	if !strings.Contains(content, "To: spark@example.com") {
		t.Errorf("missing To header in MIME: %s", content)
	}
	if !strings.Contains(content, "X-SparkWA-Request-ID: uuid-123") {
		t.Errorf("missing X-SparkWA-Request-ID header in MIME: %s", content)
	}
	if !strings.Contains(content, "X-SparkWA-Sender: +19876543210") {
		t.Errorf("missing X-SparkWA-Sender header in MIME: %s", content)
	}

	// Verify that body is valid base64
	lines := strings.Split(content, "\r\n\r\n")
	if len(lines) < 2 {
		t.Fatalf("expected header and body separated by double CRLF")
	}
	encodedBody := strings.TrimSpace(lines[1])
	decoded, err := base64.StdEncoding.DecodeString(encodedBody)
	if err != nil {
		t.Fatalf("failed to decode base64 body: %v", err)
	}
	if string(decoded) != body {
		t.Errorf("decoded body mismatch: got %q, want %q", string(decoded), body)
	}
}

func TestBuildRFC2822InvalidAddresses(t *testing.T) {
	_, err := BuildRFC2822Message("invalid-email", "valid@example.com", "sub", "body", "id", "+100")
	if err == nil {
		t.Errorf("expected error for invalid 'from' address")
	}

	_, err = BuildRFC2822Message("valid@example.com", "invalid-to", "sub", "body", "id", "+100")
	if err == nil {
		t.Errorf("expected error for invalid 'to' address")
	}
}

func TestStripHTML(t *testing.T) {
	html := "<div><p>Hello <b>World</b>!</p>&nbsp;<span>This is &lt;tested&gt; &amp; verified.</span></div>"
	got := StripHTML(html)
	want := "Hello  World !   This is <tested> & verified."
	if got != want {
		t.Errorf("StripHTML() = %q, want %q", got, want)
	}
}

func TestExtractBodyText(t *testing.T) {
	plainText := "Here is the response from Gemini Spark."
	b64Plain := base64.URLEncoding.EncodeToString([]byte(plainText))

	payload := &gmail.MessagePart{
		MimeType: "multipart/alternative",
		Parts: []*gmail.MessagePart{
			{
				MimeType: "text/plain",
				Body: &gmail.MessagePartBody{
					Data: b64Plain,
				},
			},
			{
				MimeType: "text/html",
				Body: &gmail.MessagePartBody{
					Data: base64.URLEncoding.EncodeToString([]byte("<p>HTML version</p>")),
				},
			},
		},
	}

	extracted := ExtractBodyText(payload)
	if extracted != plainText {
		t.Errorf("ExtractBodyText() = %q, want %q", extracted, plainText)
	}
}

func TestTokenSaveAndRead(t *testing.T) {
	tempDir := t.TempDir()
	tokenPath := filepath.Join(tempDir, "token.json")

	originalTok := &oauth2.Token{
		AccessToken:  "mock-access-token-123",
		TokenType:    "Bearer",
		RefreshToken: "mock-refresh-token-456",
		Expiry:       time.Now().Add(1 * time.Hour).UTC(),
	}

	if err := SaveToken(tokenPath, originalTok); err != nil {
		t.Fatalf("SaveToken failed: %v", err)
	}

	// Verify file permissions 0600
	info, err := os.Stat(tokenPath)
	if err != nil {
		t.Fatalf("failed to stat token file: %v", err)
	}
	mode := info.Mode().Perm()
	if mode != 0600 {
		t.Errorf("expected token file mode 0600, got %o", mode)
	}

	// Read token back
	readTok, err := ReadSavedToken(tokenPath)
	if err != nil {
		t.Fatalf("ReadSavedToken failed: %v", err)
	}

	if readTok.AccessToken != originalTok.AccessToken || readTok.RefreshToken != originalTok.RefreshToken {
		t.Errorf("token data mismatch: got %+v, want %+v", readTok, originalTok)
	}
}
