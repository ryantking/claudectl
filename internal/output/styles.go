package output

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// SuccessStyle styles success messages.
	SuccessStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))

	// ErrorStyle styles error messages.
	ErrorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))

	// InfoStyle styles info messages.
	InfoStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("4"))

	// BoldStyle styles bold text.
	BoldStyle = lipgloss.NewStyle().Bold(true)
)
