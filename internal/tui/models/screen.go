package models

import tea "github.com/charmbracelet/bubbletea"

// Screen represents a TUI screen/page
type Screen interface {
	Init() tea.Cmd
	Update(msg tea.Msg) (Screen, tea.Cmd)
	View() string
	SetFocus(hasFocus bool)
}