package tui

import (
	"os"

	"github.com/charmbracelet/lipgloss"
)

// Renderer targets stderr so colors work even when stdout is captured by shell wrapper.
var Renderer = lipgloss.NewRenderer(os.Stderr)

var (
	Green  = Renderer.NewStyle().Foreground(lipgloss.Color("2"))
	Red    = Renderer.NewStyle().Foreground(lipgloss.Color("1"))
	Yellow = Renderer.NewStyle().Foreground(lipgloss.Color("3"))
	Cyan   = Renderer.NewStyle().Foreground(lipgloss.Color("6"))
	Bold   = Renderer.NewStyle().Bold(true)
	Dim    = Renderer.NewStyle().Faint(true)

	Selected   = Renderer.NewStyle().Foreground(lipgloss.Color("6")).Bold(true)
	Unselected = Renderer.NewStyle().Foreground(lipgloss.Color("7"))
	Cursor     = Renderer.NewStyle().Foreground(lipgloss.Color("6"))
)
