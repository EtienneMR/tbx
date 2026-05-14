package tui

import (
	"fmt"
	"os"

	"charm.land/lipgloss/v2"
)

var (
	colorInfo    = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))  // bright blue
	colorSuccess = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))  // bright green
	colorWarn    = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))  // bright yellow
	colorError   = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))   // bright red
	colorMuted   = lipgloss.NewStyle().Foreground(lipgloss.Color("240")) // grey
	colorBold    = lipgloss.NewStyle().Bold(true)

	borderStyle = colorMuted.PaddingLeft(1).BorderStyle(lipgloss.ThickBorder()).BorderLeft(true).BorderForeground(lipgloss.Color("240"))
)

const (
	iconStep    = "◆"
	iconSuccess = "✔"
	iconWarn    = "⚠"
	iconError   = "✘"
	iconInfo    = "ℹ"
	iconRun     = "▶"
)

// Verbosity controls whether Debug lines are printed.
var Verbosity int // 0 = normal, 1 = verbose

// Step prints a primary action line (blue ◆).
func Step(format string, a ...any) {
	print(colorInfo, iconStep, format, a...)
}

// Success prints a success confirmation (green ✔).
func Success(format string, a ...any) {
	print(colorSuccess, iconSuccess, format, a...)
}

// Warn prints a non-fatal warning (yellow ⚠).
func Warn(format string, a ...any) {
	print(colorWarn, iconWarn, format, a...)
}

// Error prints an error line (red ✘) without exiting.
func Error(format string, a ...any) {
	eprint(colorError, iconError, format, a...)
}

// Info prints a neutral informational line (blue ℹ).
func Info(format string, a ...any) {
	print(colorInfo, iconInfo, format, a...)
}

// Run prints the command that is about to be executed (grey ▶).
// Only shown when Verbosity >= 1.
func Run(format string, a ...any) {
	if Verbosity < 1 {
		return
	}
	print(colorMuted, iconRun, format, a...)
}

// Debug prints a grey debug line. Only shown when Verbosity >= 2.
func Debug(format string, a ...any) {
	if Verbosity < 2 {
		return
	}
	print(colorMuted, iconInfo, format, a...)
}

// Blank prints an empty line.
func Blank() { fmt.Fprintln(os.Stdout) }

// Header prints a bold section title with a top margin.
func Header(title string) {
	fmt.Fprintln(os.Stdout)
	fmt.Fprintln(os.Stdout, colorBold.Render(title))
}

// Fatal prints an error line then exits with code 1.
func Fatal(format string, a ...any) {
	eprint(colorError, iconError, format, a...)
	os.Exit(1)
}

// Check call Fatal if an error is provided
func Check(err error) {
	if err != nil {
		Fatal("%v", err)
	}
}

func print(style lipgloss.Style, icon, format string, a ...any) {
	prefix := style.Render(icon)
	msg := fmt.Sprintf(format, a...)
	fmt.Fprintf(os.Stdout, "%s %s\n", prefix, msg)
}

func eprint(style lipgloss.Style, icon, format string, a ...any) {
	prefix := style.Render(icon)
	msg := fmt.Sprintf(format, a...)
	fmt.Fprintf(os.Stderr, "%s %s\n", prefix, msg)
}

// Indent wraps multi-line output (e.g. command stdout) with a grey left margin.
func Indent(block string) {
	fmt.Fprintln(os.Stdout, borderStyle.Render(block))
}
