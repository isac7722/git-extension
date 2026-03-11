package tui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// PromptModel handles text input.
type PromptModel struct {
	input    textinput.Model
	label    string
	quit     bool
	done     bool
}

// NewPrompt creates a new text prompt.
func NewPrompt(label, placeholder string) PromptModel {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.Focus()
	return PromptModel{
		input: ti,
		label: label,
	}
}

func (m PromptModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m PromptModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			m.done = true
			return m, tea.Quit
		case "esc", "ctrl+c":
			m.quit = true
			return m, tea.Quit
		}
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m PromptModel) View() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "%s %s\n", Cyan.Render("?"), Bold.Render(m.label))
	sb.WriteString(m.input.View())
	return sb.String()
}

// Value returns the entered text, or empty string if cancelled.
func (m PromptModel) Value() string {
	if m.quit {
		return ""
	}
	return m.input.Value()
}

// RunPromptWithValue runs a text prompt with a pre-filled value and returns the entered value.
func RunPromptWithValue(label, placeholder, value string) (string, bool, error) {
	m := NewPrompt(label, placeholder)
	m.input.SetValue(value)
	p := tea.NewProgram(m, tea.WithOutput(os.Stderr))
	finalModel, err := p.Run()
	if err != nil {
		return "", false, err
	}
	result := finalModel.(PromptModel)
	if result.quit {
		return "", false, nil
	}
	return result.Value(), true, nil
}

// RunPrompt runs a text prompt and returns the entered value.
func RunPrompt(label, placeholder string) (string, bool, error) {
	m := NewPrompt(label, placeholder)
	p := tea.NewProgram(m, tea.WithOutput(os.Stderr))
	finalModel, err := p.Run()
	if err != nil {
		return "", false, err
	}
	result := finalModel.(PromptModel)
	if result.quit {
		return "", false, nil
	}
	return result.Value(), true, nil
}
