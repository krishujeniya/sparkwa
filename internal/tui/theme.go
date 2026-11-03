package tui

import (
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

// Monochromatic ANSI Styling Tokens (Black and White Only).
// No chromatic hues (red, green, blue, yellow) are used in the interface.
const (
	Reset       = "\033[0m"
	BoldText    = "\033[1m"
	DimText     = "\033[2m"
	InverseText = "\033[7m"
	Underline   = "\033[4m"
)

// Bold wraps text in high-intensity bold white.
func Bold(s string) string {
	return BoldText + s + Reset
}

// Dim wraps text in subtle low-intensity/muted gray.
func Dim(s string) string {
	return DimText + s + Reset
}

// Inverse wraps text in inverted high-contrast black-on-white.
func Inverse(s string) string {
	return InverseText + s + Reset
}

// Badge returns an inverted, prominent high-contrast pill badge.
func Badge(s string) string {
	return InverseText + " " + s + " " + Reset
}

// IndicatorActive returns a bold positive monochrome status token.
func IndicatorActive(label string) string {
	return Bold("[● " + label + "]")
}

// IndicatorPending returns a dim unconfigured/pending monochrome status token.
func IndicatorPending(label string) string {
	return Dim("[○ " + label + "]")
}

// IndicatorWarn returns an attention monochrome status token.
func IndicatorWarn(label string) string {
	return Inverse("[! " + label + "]")
}

// ClearScreen clears the terminal screen and resets the cursor to (0,0).
func ClearScreen(w io.Writer) {
	fmt.Fprint(w, "\033[2J\033[H")
}

// GetTerminalWidth returns the current terminal column width, defaulting to 80.
func GetTerminalWidth() int {
	fd := int(os.Stdout.Fd())
	if term.IsTerminal(fd) {
		w, _, err := term.GetSize(fd)
		if err == nil && w > 20 {
			return w
		}
	}
	return 80
}

// Box creates an elegant, high-contrast monochrome framed box.
func Box(title string, lines []string, width int) string {
	if width < 40 {
		width = 40
	}
	if width > 84 {
		width = 84
	}

	var sb strings.Builder
	innerWidth := width - 4

	// Top Border with embedded Title
	if title != "" {
		displayTitle := " " + strings.TrimSpace(title) + " "
		remaining := innerWidth - len(displayTitle)
		if remaining < 0 {
			remaining = 0
		}
		sb.WriteString("┌─" + displayTitle + strings.Repeat("─", remaining) + "┐\n")
	} else {
		sb.WriteString("┌" + strings.Repeat("─", innerWidth+2) + "┐\n")
	}

	// Content Lines
	for _, line := range lines {
		// Calculate visible length without ANSI codes for proper padding
		visLen := visibleLength(line)
		pad := innerWidth - visLen
		if pad < 0 {
			pad = 0
		}
		sb.WriteString("│ " + line + strings.Repeat(" ", pad) + " │\n")
	}

	// Bottom Border
	sb.WriteString("└" + strings.Repeat("─", innerWidth+2) + "┘\n")
	return sb.String()
}

// visibleLength estimates printable character width by stripping standard ANSI escape sequences.
func visibleLength(s string) int {
	inEscape := false
	length := 0
	for _, r := range s {
		if r == '\033' {
			inEscape = true
			continue
		}
		if inEscape {
			if r == 'm' || r == 'J' || r == 'H' || r == 'K' {
				inEscape = false
			}
			continue
		}
		length++
	}
	return length
}
