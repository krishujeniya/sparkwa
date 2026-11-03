package tui

import (
	"fmt"
	"io"
	"strings"
)

// PrintLogo renders the official monochromatic SparkWA logo.
// If the terminal column width is under 80 columns, it automatically renders
// the sleek, compact horizontal banner instead to prevent line wrapping.
func PrintLogo(w io.Writer, termWidth int) {
	if termWidth < 80 {
		printCompactLogo(w)
		return
	}
	printFullLogo(w)
}

func printFullLogo(w io.Writer) {
	banner := `
╔════════════════════════════════════════════════════════════════════════════════╗
║  ███████╗██████╗  █████╗ ██████╗ ██╗  ██╗   ██╗    ██╗ █████╗                  ║
║  ██╔════╝██╔══██╗██╔══██╗██╔══██╗██║ ██╔╝   ██║    ██║██╔══██╗                 ║
║  ███████╗██████╔╝███████║██████╔╝█████╔╝    ██║ █╗ ██║███████║   ⚡             ║
║  ╚════██║██╔═══╝ ██╔══██║██╔══██╗██╔═██╗    ██║███╗██║██╔══██║                 ║
║  ███████║██║     ██║  ██║██║  ██║██║  ██╗   ╚███╔███╔╝██║  ██║                 ║
║  ╚══════╝╚═╝     ╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═╝    ╚══╝╚══╝ ╚═╝  ╚═╝                 ║
║                                                                                ║
║  THE HEADLESS WHATSAPP TO GEMINI SPARK BRIDGE                                  ║
║  Pure Go • Single Native Binary • Zero Dependencies • Production Hardened      ║
╚════════════════════════════════════════════════════════════════════════════════╝`

	lines := strings.Split(strings.Trim(banner, "\n"), "\n")
	for i, l := range lines {
		if i == 0 || i == len(lines)-1 {
			fmt.Fprintln(w, Bold(l))
		} else if strings.Contains(l, "THE HEADLESS") {
			fmt.Fprintln(w, Bold(l))
		} else if strings.Contains(l, "Pure Go") {
			fmt.Fprintln(w, Dim(l))
		} else {
			fmt.Fprintln(w, l)
		}
	}
}

func printCompactLogo(w io.Writer) {
	compact := `
┌────────────────────────────────────────────────────────┐
│  ⚡ SPARKWA │ WhatsApp ⇄ Google Gemini Spark Bridge    │
│  Pure Go • Single Binary • Zero External Dependencies  │
└────────────────────────────────────────────────────────┘`

	lines := strings.Split(strings.Trim(compact, "\n"), "\n")
	for i, l := range lines {
		if i == 1 {
			fmt.Fprintln(w, Bold(l))
		} else if i == 2 {
			fmt.Fprintln(w, Dim(l))
		} else {
			fmt.Fprintln(w, l)
		}
	}
}
