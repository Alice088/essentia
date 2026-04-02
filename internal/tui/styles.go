package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// Colors inspired by Claude Code orange theme
var (
	Orange   = lipgloss.Color("#FF6B35")
	DarkGray = lipgloss.Color("#2D3748")
	LightGray = lipgloss.Color("#4A5568")
	White     = lipgloss.Color("#FFFFFF")
)

var styles = struct {
	Tab       lipgloss.Style
	ActiveTab lipgloss.Style
	Title     lipgloss.Style
	Footer    lipgloss.Style
	Content   lipgloss.Style
}{
	Tab: lipgloss.NewStyle().
		Padding(0, 2).
		Margin(0, 1).
		Foreground(LightGray).
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(DarkGray),

	ActiveTab: lipgloss.NewStyle().
		Padding(0, 2).
		Margin(0, 1).
		Foreground(Orange).
		Bold(true).
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(Orange),

	Title: lipgloss.NewStyle().
		Foreground(Orange).
		Bold(true).
		Align(lipgloss.Center).
		MarginTop(1).
		MarginBottom(1),

	Footer: lipgloss.NewStyle().
		Foreground(LightGray).
		Padding(0, 1).
		Border(lipgloss.NormalBorder(), true, false, false, false).
		BorderForeground(DarkGray).
		Width(0),

	Content: lipgloss.NewStyle().
		Padding(1, 2),
}