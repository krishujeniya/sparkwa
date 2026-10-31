package message

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestSplitTextShort(t *testing.T) {
	short := "Hello world, this is a short test message."
	chunks := SplitText(short, 100)
	if len(chunks) != 1 {
		t.Fatalf("expected 1 chunk, got %d", len(chunks))
	}
	if chunks[0] != short {
		t.Errorf("expected %q, got %q", short, chunks[0])
	}
}

func TestSplitTextMultiChunk(t *testing.T) {
	// Build a text of 5 lines, each 50 chars => 250 chars
	line := "This is a line of text that has exactly 48 chars\n"
	text := strings.Repeat(line, 5)

	// Split with maxChunkSize = 100
	chunks := SplitText(text, 100)
	if len(chunks) < 3 {
		t.Fatalf("expected at least 3 chunks, got %d", len(chunks))
	}

	for i, chunk := range chunks {
		if !strings.HasPrefix(chunk, "[") {
			t.Errorf("chunk %d missing multi-part header: %s", i, chunk)
		}
	}
}

func TestSplitTextLongUnbrokenWord(t *testing.T) {
	longWord := strings.Repeat("A", 150)
	chunks := SplitText(longWord, 50)
	if len(chunks) != 3 {
		t.Fatalf("expected 3 chunks for 150 chars with limit 50, got %d", len(chunks))
	}
}

func TestSplitTextUnicodeAndEmojis(t *testing.T) {
	emojiText := strings.Repeat("⚡🚀🤖 SparkWA Lite Multi-byte text test! \n", 20)
	chunks := SplitText(emojiText, 150)

	for i, c := range chunks {
		if !utf8.ValidString(c) {
			t.Errorf("chunk %d contains invalid UTF-8 bytes", i)
		}
	}
}

func TestFormatters(t *testing.T) {
	ack := FormatSubmissionAck("req-123", "Test command")
	if !strings.Contains(ack, "req-123") || !strings.Contains(ack, "SUBMITTED") {
		t.Errorf("unexpected submission ack: %s", ack)
	}

	comp := FormatCompletion("req-123", "Finished result")
	if !strings.Contains(comp, "COMPLETED") || !strings.Contains(comp, "Finished result") {
		t.Errorf("unexpected completion: %s", comp)
	}

	help := FormatHelp()
	if !strings.Contains(help, "/status") || !strings.Contains(help, "/jobs") {
		t.Errorf("unexpected help: %s", help)
	}
}
