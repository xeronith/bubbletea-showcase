package app

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func sizedModel(t *testing.T) *Model {
	t.Helper()
	m := New()
	m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	return m
}

func press(m *Model, key tea.KeyPressMsg) {
	m.Update(key)
}

func runeKey(r rune) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: r, Text: string(r)}
}

func TestHasFivePages(t *testing.T) {
	m := New()
	if got, want := len(m.pages), 5; got != want {
		t.Fatalf("page count = %d, want %d", got, want)
	}
}

func TestNumberKeysSwitchPages(t *testing.T) {
	m := sizedModel(t)

	press(m, runeKey('3'))
	if m.current != 2 {
		t.Fatalf("current = %d, want 2", m.current)
	}

	press(m, runeKey('1'))
	if m.current != 0 {
		t.Fatalf("current = %d, want 0", m.current)
	}
}

func TestTabAdvancesExceptOnFormPages(t *testing.T) {
	m := sizedModel(t)

	press(m, tea.KeyPressMsg{Code: tea.KeyTab})
	if m.current != 1 {
		t.Fatalf("tab from overview: current = %d, want 1", m.current)
	}

	// Kitchen consumes tab for field navigation.
	press(m, tea.KeyPressMsg{Code: tea.KeyTab})
	if m.current != 1 {
		t.Fatalf("tab on kitchen switched page: current = %d, want 1", m.current)
	}

	press(m, runeKey('4'))
	press(m, tea.KeyPressMsg{Code: tea.KeyTab})
	if m.current != 4 {
		t.Fatalf("tab from motion: current = %d, want 4", m.current)
	}
}

func TestHelpOverlayToggles(t *testing.T) {
	m := sizedModel(t)

	press(m, runeKey('?'))
	if !m.showHelp {
		t.Fatal("help should be open")
	}
	press(m, runeKey('x'))
	if m.showHelp {
		t.Fatal("any key should close help")
	}
}

func TestSidebarClickSwitchesPage(t *testing.T) {
	m := sizedModel(t)
	if !m.hasSidebar() {
		t.Fatal("expected sidebar at 120 columns")
	}

	msg := tea.MouseClickMsg{
		X:      8,
		Y:      mainTop + sidebarItem + 3,
		Button: tea.MouseLeft,
	}
	m.Update(msg)
	if m.current != 3 {
		t.Fatalf("current = %d, want 3", m.current)
	}
}

func TestEditingPageKeepsKeys(t *testing.T) {
	m := sizedModel(t)
	press(m, runeKey('2'))

	// Move focus from Tea to Name. Kitchen owns Tab.
	press(m, tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift})
	if m.current != 1 {
		t.Fatalf("current = %d, want 1", m.current)
	}

	// Digits are typed into the name field, not treated as navigation.
	press(m, runeKey('3'))
	if m.current != 1 {
		t.Fatalf("digit while editing switched page: current = %d", m.current)
	}
}

func TestViewFillsTheTerminal(t *testing.T) {
	m := sizedModel(t)
	v := m.View()
	if got, want := lipgloss.Width(v.Content), 120; got != want {
		t.Fatalf("width = %d, want %d", got, want)
	}
	if got, want := lipgloss.Height(v.Content), 40; got != want {
		t.Fatalf("height = %d, want %d", got, want)
	}
	if !v.AltScreen {
		t.Fatal("showcase should use the alternate screen")
	}
}

func TestTooSmallTerminalStillRenders(t *testing.T) {
	m := New()
	m.Update(tea.WindowSizeMsg{Width: 60, Height: 20})
	v := m.View()
	if got, want := lipgloss.Width(v.Content), 60; got != want {
		t.Fatalf("width = %d, want %d", got, want)
	}
	if got, want := lipgloss.Height(v.Content), 20; got != want {
		t.Fatalf("height = %d, want %d", got, want)
	}
}
