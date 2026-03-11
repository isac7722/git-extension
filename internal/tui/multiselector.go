package tui

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// MultiItem represents a checkable item.
type MultiItem struct {
	Label   string
	Value   string
	Hint    string
	Checked bool
}

// MultiSelectorModel is a multi-select list with checkboxes.
type MultiSelectorModel struct {
	Items  []MultiItem
	cursor int
	quit   bool
	done   bool
}

// NewMultiSelector creates a new multi-selector.
func NewMultiSelector(items []MultiItem) MultiSelectorModel {
	return MultiSelectorModel{Items: items}
}

func (m MultiSelectorModel) Init() tea.Cmd { return nil }

func (m MultiSelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch keyMsg.String() {
	case "up", "k":
		m.moveCursor(-1)
	case "down", "j":
		m.moveCursor(1)
	case " ":
		m.Items[m.cursor].Checked = !m.Items[m.cursor].Checked
	case "a":
		toggleAll(m.Items)
	case "enter":
		m.done = true
		return m, tea.Quit
	case "esc", "q", "ctrl+c":
		m.quit = true
		return m, tea.Quit
	}
	return m, nil
}

func (m *MultiSelectorModel) moveCursor(delta int) {
	next := m.cursor + delta
	if next >= 0 && next < len(m.Items) {
		m.cursor = next
	}
}

func (m MultiSelectorModel) View() string {
	var sb strings.Builder

	sb.WriteString(Dim.Render("Select branches to delete:") + "\n\n")

	for i, item := range m.Items {
		cursor := "  "
		if i == m.cursor {
			cursor = Cursor.Render("❯ ")
		}

		check := "[ ]"
		if item.Checked {
			check = Green.Render("[✔]")
		}

		label := item.Label
		if i == m.cursor {
			label = Bold.Render(label)
		}

		hint := ""
		if item.Hint != "" {
			hint = "  " + Dim.Render(item.Hint)
		}

		fmt.Fprintf(&sb, "%s%s %s%s\n", cursor, check, label, hint)
	}

	count := 0
	for _, item := range m.Items {
		if item.Checked {
			count++
		}
	}

	fmt.Fprintf(&sb, "\n%s\n", Dim.Render(fmt.Sprintf(
		"↑↓/jk: move  ␣: toggle  a: all  ⏎: confirm (%d selected)  esc/q: cancel", count,
	)))

	return sb.String()
}

func toggleAll(items []MultiItem) {
	allChecked := true
	for _, item := range items {
		if !item.Checked {
			allChecked = false
			break
		}
	}
	for i := range items {
		items[i].Checked = !allChecked
	}
}

// SelectedIndices returns indices of checked items, or nil if cancelled.
func (m MultiSelectorModel) SelectedIndices() []int {
	if m.quit {
		return nil
	}
	var indices []int
	for i, item := range m.Items {
		if item.Checked {
			indices = append(indices, i)
		}
	}
	return indices
}

// RunMultiSelector runs the multi-selector and returns selected indices.
func RunMultiSelector(items []MultiItem) ([]int, error) {
	m := NewMultiSelector(items)
	p := tea.NewProgram(m, tea.WithOutput(os.Stderr))
	finalModel, err := p.Run()
	if err != nil {
		return nil, err
	}
	return finalModel.(MultiSelectorModel).SelectedIndices(), nil
}
