package tui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

// TextAreaModel handles multi-line text input.
type TextAreaModel struct {
	textarea textarea.Model
	label    string
	quit     bool
	done     bool
}

// NewTextArea creates a new multi-line text area.
func NewTextArea(label, placeholder string) TextAreaModel {
	ta := textarea.New()
	ta.Placeholder = placeholder
	ta.ShowLineNumbers = false
	ta.SetWidth(72)
	ta.SetHeight(5)
	ta.KeyMap.InsertNewline.SetKeys("alt+enter")
	ta.Focus()
	return TextAreaModel{
		textarea: ta,
		label:    label,
	}
}

func (m TextAreaModel) Init() tea.Cmd {
	return textarea.Blink
}

func (m TextAreaModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			m.done = true
			return m, tea.Quit
		case "alt+enter":
			m.textarea.InsertString("\n")
			return m, nil
		case "esc", "ctrl+c":
			m.quit = true
			return m, tea.Quit
		}
	}
	var cmd tea.Cmd
	m.textarea, cmd = m.textarea.Update(msg)
	return m, cmd
}

func (m TextAreaModel) View() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "%s %s %s\n", Cyan.Render("?"), Bold.Render(m.label), Dim.Render("(Enter to submit, Alt(Option)+Enter for newline)"))
	sb.WriteString(m.textarea.View())
	return sb.String()
}

// Value returns the entered text, or empty string if cancelled.
func (m TextAreaModel) Value() string {
	if m.quit {
		return ""
	}
	return m.textarea.Value()
}

// RunTextArea runs a multi-line text area and returns the entered value.
func RunTextArea(label, placeholder string) (string, bool, error) {
	m := NewTextArea(label, placeholder)
	p := tea.NewProgram(m, tea.WithOutput(os.Stderr))
	finalModel, err := p.Run()
	if err != nil {
		return "", false, err
	}
	result := finalModel.(TextAreaModel)
	if result.quit {
		return "", false, nil
	}
	return result.Value(), true, nil
}
