package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func testItems() []MultiItem {
	return []MultiItem{
		{Label: "branch-a", Value: "a", Hint: "local"},
		{Label: "branch-b", Value: "b", Hint: "remote"},
		{Label: "branch-c", Value: "c", Hint: "local + remote"},
	}
}

func TestNewMultiSelector(t *testing.T) {
	items := testItems()
	m := NewMultiSelector(items)

	if len(m.Items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(m.Items))
	}
	if m.cursor != 0 {
		t.Error("cursor should start at 0")
	}
	if m.quit || m.done {
		t.Error("quit and done should be false initially")
	}
}

func TestMultiSelectorInit(t *testing.T) {
	m := NewMultiSelector(testItems())
	cmd := m.Init()
	if cmd != nil {
		t.Error("Init should return nil")
	}
}

func TestMultiSelector_CursorMovement(t *testing.T) {
	m := NewMultiSelector(testItems())

	// Move down
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = result.(MultiSelectorModel)
	if m.cursor != 1 {
		t.Errorf("expected cursor at 1, got %d", m.cursor)
	}

	// Move down with arrow
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = result.(MultiSelectorModel)
	if m.cursor != 2 {
		t.Errorf("expected cursor at 2, got %d", m.cursor)
	}

	// Can't go below bounds
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = result.(MultiSelectorModel)
	if m.cursor != 2 {
		t.Errorf("cursor should stay at 2, got %d", m.cursor)
	}

	// Move up
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = result.(MultiSelectorModel)
	if m.cursor != 1 {
		t.Errorf("expected cursor at 1, got %d", m.cursor)
	}

	// Move up with arrow
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = result.(MultiSelectorModel)
	if m.cursor != 0 {
		t.Errorf("expected cursor at 0, got %d", m.cursor)
	}

	// Can't go above bounds
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = result.(MultiSelectorModel)
	if m.cursor != 0 {
		t.Errorf("cursor should stay at 0, got %d", m.cursor)
	}
}

func TestMultiSelector_Toggle(t *testing.T) {
	m := NewMultiSelector(testItems())

	// Toggle first item
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	m = result.(MultiSelectorModel)
	if !m.Items[0].Checked {
		t.Error("item 0 should be checked after toggle")
	}

	// Toggle again to uncheck
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	m = result.(MultiSelectorModel)
	if m.Items[0].Checked {
		t.Error("item 0 should be unchecked after second toggle")
	}
}

func TestMultiSelector_ToggleAll(t *testing.T) {
	m := NewMultiSelector(testItems())

	// Toggle all on
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m = result.(MultiSelectorModel)
	for i, item := range m.Items {
		if !item.Checked {
			t.Errorf("item %d should be checked after toggle all", i)
		}
	}

	// Toggle all off
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m = result.(MultiSelectorModel)
	for i, item := range m.Items {
		if item.Checked {
			t.Errorf("item %d should be unchecked after second toggle all", i)
		}
	}
}

func TestMultiSelector_Enter(t *testing.T) {
	m := NewMultiSelector(testItems())

	// Check first item
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	m = result.(MultiSelectorModel)

	// Press enter
	result, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = result.(MultiSelectorModel)
	if !m.done {
		t.Error("done should be true after enter")
	}
	if cmd == nil {
		t.Error("enter should return a Quit command")
	}

	indices := m.SelectedIndices()
	if len(indices) != 1 || indices[0] != 0 {
		t.Errorf("expected [0], got %v", indices)
	}
}

func TestMultiSelector_Cancel(t *testing.T) {
	tests := []struct {
		name string
		key  tea.KeyMsg
	}{
		{"esc", tea.KeyMsg{Type: tea.KeyEsc}},
		{"q", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}},
		{"ctrl+c", tea.KeyMsg{Type: tea.KeyCtrlC}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMultiSelector(testItems())
			// Check an item first
			result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
			m = result.(MultiSelectorModel)

			result, cmd := m.Update(tt.key)
			m = result.(MultiSelectorModel)
			if !m.quit {
				t.Error("quit should be true")
			}
			if cmd == nil {
				t.Error("cancel should return a Quit command")
			}

			// SelectedIndices returns nil when cancelled
			if m.SelectedIndices() != nil {
				t.Error("SelectedIndices should return nil when cancelled")
			}
		})
	}
}

func TestMultiSelector_SelectedIndices_NoneSelected(t *testing.T) {
	m := NewMultiSelector(testItems())
	// Press enter without selecting anything
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = result.(MultiSelectorModel)

	indices := m.SelectedIndices()
	if indices != nil {
		t.Errorf("expected nil when none selected, got %v", indices)
	}
}

func TestMultiSelector_SelectedIndices_Multiple(t *testing.T) {
	m := NewMultiSelector(testItems())

	// Check item 0
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	m = result.(MultiSelectorModel)

	// Move to item 2 and check
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = result.(MultiSelectorModel)
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = result.(MultiSelectorModel)
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	m = result.(MultiSelectorModel)

	// Confirm
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = result.(MultiSelectorModel)

	indices := m.SelectedIndices()
	if len(indices) != 2 {
		t.Fatalf("expected 2 selected, got %d", len(indices))
	}
	if indices[0] != 0 || indices[1] != 2 {
		t.Errorf("expected [0, 2], got %v", indices)
	}
}

func TestMultiSelector_View(t *testing.T) {
	m := NewMultiSelector(testItems())
	view := m.View()

	// Should contain the header
	if !strings.Contains(view, "Select branches to delete:") {
		t.Error("view should contain header")
	}

	// Should contain item labels
	if !strings.Contains(view, "branch-a") {
		t.Error("view should contain branch-a")
	}
	if !strings.Contains(view, "branch-b") {
		t.Error("view should contain branch-b")
	}

	// Should contain help text
	if !strings.Contains(view, "toggle") {
		t.Error("view should contain help text")
	}

	// Should show 0 selected initially
	if !strings.Contains(view, "0 selected") {
		t.Error("view should show 0 selected")
	}
}

func TestMultiSelector_View_WithChecked(t *testing.T) {
	items := testItems()
	items[0].Checked = true
	m := NewMultiSelector(items)
	view := m.View()

	if !strings.Contains(view, "1 selected") {
		t.Error("view should show 1 selected")
	}
}

func TestMultiSelector_NonKeyMsg(t *testing.T) {
	m := NewMultiSelector(testItems())
	// Sending a non-key message should not change state
	result, cmd := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m2 := result.(MultiSelectorModel)
	if cmd != nil {
		t.Error("non-key message should return nil cmd")
	}
	if m2.cursor != 0 {
		t.Error("cursor should not change on non-key message")
	}
}

func TestToggleAll_Standalone(t *testing.T) {
	items := []MultiItem{
		{Label: "a", Checked: false},
		{Label: "b", Checked: false},
	}

	// All unchecked -> all checked
	toggleAll(items)
	for i, item := range items {
		if !item.Checked {
			t.Errorf("item %d should be checked", i)
		}
	}

	// All checked -> all unchecked
	toggleAll(items)
	for i, item := range items {
		if item.Checked {
			t.Errorf("item %d should be unchecked", i)
		}
	}

	// Partial checked -> all checked
	items[0].Checked = true
	toggleAll(items)
	for i, item := range items {
		if !item.Checked {
			t.Errorf("item %d should be checked (from partial)", i)
		}
	}
}
