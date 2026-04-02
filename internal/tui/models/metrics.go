package models

import (
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// MetricsScreen represents the metrics screen
type MetricsScreen struct {
	width  int
	height int
}

// NewMetricsScreen creates a new MetricsScreen
func NewMetricsScreen() Screen {
	return &MetricsScreen{}
}

// Init initializes the metrics screen
func (m *MetricsScreen) Init() tea.Cmd {
	return nil
}

// Update handles messages for the metrics screen
func (m *MetricsScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

// View renders the metrics screen
func (m *MetricsScreen) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF6B35")).
		Bold(true).
		Align(lipgloss.Center).
		MarginTop(1).
		MarginBottom(1)

	contentStyle := lipgloss.NewStyle().
		Padding(1, 2)

	title := titleStyle.Render("Metrics Dashboard")
	message := "This screen will show:\n" +
		"• Key metrics from Prometheus\n" +
		"• Metrics grouped by category (processing, memory, errors)\n" +
		"• Auto-refresh every 5 seconds\n" +
		"• Timestamp of last update\n\n" +
		"To be implemented in Phase 4"

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		contentStyle.Render(message),
	)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}