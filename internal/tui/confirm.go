package tui

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// ConfirmModel handles y/N confirmation.
type ConfirmModel struct {
	message string
	result  bool
	done    bool
}

// NewConfirm creates a confirmation prompt.
func NewConfirm(message string) ConfirmModel {
	return ConfirmModel{message: message}
}

func (m ConfirmModel) Init() tea.Cmd { return nil }

func (m ConfirmModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch strings.ToLower(msg.String()) {
		case "y":
			m.result = true
			m.done = true
			return m, tea.Quit
		case "n", "esc", "ctrl+c", "enter":
			m.result = false
			m.done = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m ConfirmModel) View() string {
	return fmt.Sprintf("%s %s %s", Yellow.Render("?"), m.message, Dim.Render("[y/N]"))
}

// Confirmed returns the result.
func (m ConfirmModel) Confirmed() bool {
	return m.result
}

// RunConfirm runs a confirmation prompt.
func RunConfirm(message string) (bool, error) {
	m := NewConfirm(message)
	p := tea.NewProgram(m, tea.WithOutput(os.Stderr), tea.WithInputTTY())
	finalModel, err := p.Run()
	if err != nil {
		return false, err
	}
	return finalModel.(ConfirmModel).Confirmed(), nil
}
