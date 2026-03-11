package tui

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// SelectorItem represents an item in the selector.
type SelectorItem struct {
	Label         string
	Value         string
	Hint          string // shown dimmed after label
	FormattedHint string // pre-formatted hint (bypasses default dim styling)
	Selected      bool   // pre-selected marker (e.g., current user)
}

// SelectorModel is a single-select list.
type SelectorModel struct {
	Items   []SelectorItem
	cursor  int
	chosen  int
	header  string
	quit    bool
}

// NewSelector creates a new selector model.
func NewSelector(items []SelectorItem, header string) SelectorModel {
	cursor := 0
	for i, item := range items {
		if item.Selected {
			cursor = i
			break
		}
	}
	return SelectorModel{
		Items:  items,
		cursor: cursor,
		chosen: -1,
		header: header,
	}
}

func (m SelectorModel) Init() tea.Cmd { return nil }

func (m SelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.Items)-1 {
				m.cursor++
			}
		case "enter":
			m.chosen = m.cursor
			return m, tea.Quit
		case "esc", "q", "ctrl+c":
			m.quit = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m SelectorModel) View() string {
	var sb strings.Builder

	if m.header != "" {
		sb.WriteString(Dim.Render(m.header) + "\n\n")
	}

	for i, item := range m.Items {
		cursor := "  "
		if i == m.cursor {
			cursor = Cursor.Render("❯ ")
		}

		label := item.Label
		if i == m.cursor {
			label = Selected.Render(label)
		}

		hint := ""
		if item.FormattedHint != "" {
			hint = " " + item.FormattedHint
		} else if item.Hint != "" {
			hint = " " + Dim.Render(item.Hint)
		}

		marker := ""
		if item.Selected {
			marker = " " + Green.Render("✔")
		}

		fmt.Fprintf(&sb, "%s%s%s%s\n", cursor, label, hint, marker)
	}

	sb.WriteString("\n" + Dim.Render("↑↓/jk: move  ⏎: select  esc/q: cancel"))

	return sb.String()
}

// Chosen returns the selected index, or -1 if cancelled.
func (m SelectorModel) Chosen() int {
	if m.quit {
		return -1
	}
	return m.chosen
}

// RunSelector runs the selector TUI and returns the chosen index.
func RunSelector(items []SelectorItem, header string) (int, error) {
	m := NewSelector(items, header)
	p := tea.NewProgram(m, tea.WithOutput(os.Stderr), tea.WithInputTTY())
	finalModel, err := p.Run()
	if err != nil {
		return -1, err
	}
	return finalModel.(SelectorModel).Chosen(), nil
}
