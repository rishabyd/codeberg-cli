package output

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	warnStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	boldStyle    = lipgloss.NewStyle().Bold(true)
	dimStyle     = lipgloss.NewStyle().Faint(true)
	urlStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("4"))
	labelStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Bold(true)
)

func Success(msg string) string {
	return successStyle.Render("✓") + " " + msg
}

func Error(msg string) string {
	return errorStyle.Render("✗") + " " + errorStyle.Render(msg)
}

func Warning(msg string) string {
	return warnStyle.Render("!") + " " + msg
}

func Bold(msg string) string {
	return boldStyle.Render(msg)
}

func Dim(msg string) string {
	return dimStyle.Render(msg)
}

func URL(msg string) string {
	return urlStyle.Render(msg)
}

func Labeled(label, value string) string {
	return fmt.Sprintf("%s %s", labelStyle.Render(strings.TrimRight(label, ":")+":"), value)
}

func StatusUp() string {
	return successStyle.Render("Up")
}

func StatusDown() string {
	return errorStyle.Render("Down")
}

func StatusUnknown() string {
	return dimStyle.Render("Unknown")
}

func Heading(msg string) string {
	return boldStyle.Render(msg)
}

func BoldSuccess(msg string) string {
	return successStyle.Bold(true).Render("✓") + " " + boldStyle.Render(msg)
}

func PrintErr(msg string) {
	fmt.Fprintln(os.Stderr, errorStyle.Render("✗ "+msg))
}
