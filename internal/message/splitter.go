package message

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// DefaultMaxChunkSize is safe under WhatsApp's 4096 character limit while leaving room for chunk prefixes.
const DefaultMaxChunkSize = 3500

// SplitText chunks a long text message into ordered parts, preserving line and word boundaries.
func SplitText(text string, maxChunkSize int) []string {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return []string{""}
	}

	if maxChunkSize <= 0 {
		maxChunkSize = DefaultMaxChunkSize
	}

	// If whole message fits, return single chunk
	if utf8.RuneCountInString(trimmed) <= maxChunkSize {
		return []string{trimmed}
	}

	var rawChunks []string
	lines := strings.Split(trimmed, "\n")
	var current strings.Builder

	for _, line := range lines {
		// If line alone exceeds maxChunkSize, break it by words or characters
		if utf8.RuneCountInString(line) > maxChunkSize {
			if current.Len() > 0 {
				rawChunks = append(rawChunks, strings.TrimSpace(current.String()))
				current.Reset()
			}
			words := strings.Fields(line)
			for _, w := range words {
				if utf8.RuneCountInString(w) > maxChunkSize {
					// Hard character split for extremely long continuous strings
					runes := []rune(w)
					for len(runes) > 0 {
						take := maxChunkSize
						if len(runes) < take {
							take = len(runes)
						}
						rawChunks = append(rawChunks, string(runes[:take]))
						runes = runes[take:]
					}
					continue
				}

				if utf8.RuneCountInString(current.String())+utf8.RuneCountInString(w)+1 > maxChunkSize {
					rawChunks = append(rawChunks, strings.TrimSpace(current.String()))
					current.Reset()
				}
				if current.Len() > 0 {
					current.WriteString(" ")
				}
				current.WriteString(w)
			}
			continue
		}

		// Check if adding this line exceeds chunk size
		additionalLen := utf8.RuneCountInString(line)
		if current.Len() > 0 {
			additionalLen++ // for newline
		}

		if utf8.RuneCountInString(current.String())+additionalLen > maxChunkSize {
			rawChunks = append(rawChunks, strings.TrimSpace(current.String()))
			current.Reset()
		}

		if current.Len() > 0 {
			current.WriteString("\n")
		}
		current.WriteString(line)
	}

	if current.Len() > 0 {
		rawChunks = append(rawChunks, strings.TrimSpace(current.String()))
	}

	// If more than 1 chunk, annotate with header (e.g. "[1/3]")
	total := len(rawChunks)
	if total <= 1 {
		return rawChunks
	}

	chunks := make([]string, total)
	for i, chunk := range rawChunks {
		chunks[i] = fmt.Sprintf("[%d/%d]\n%s", i+1, total, chunk)
	}

	return chunks
}
