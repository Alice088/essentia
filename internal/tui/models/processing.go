package models

import (
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ProcessingScreen represents the processing screen
type ProcessingScreen struct {
	width  int
	height int
}

// NewProcessingScreen creates a new ProcessingScreen
func NewProcessingScreen() Screen {
	return &ProcessingScreen{}
}

// Init initializes the processing screen
func (p *ProcessingScreen) Init() tea.Cmd {
	return nil
}

// Update handles messages for the processing screen
func (p *ProcessingScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		p.width = msg.Width
		p.height = msg.Height
	}
	return p, nil
}

// View renders the processing screen
func (p *ProcessingScreen) View() string {
	if p.width == 0 || p.height == 0 {
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

	title := titleStyle.Render("Processing Pipeline")
	message := "This screen will show:\n" +
		"• List of jobs in the pipeline\n" +
		"• Status of each job (waiting, processing, done, error)\n" +
		"• Detailed information about selected job\n" +
		"• Auto-refresh every 2 seconds\n\n" +
		"To be implemented in Phase 3"

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		contentStyle.Render(message),
	)

	return lipgloss.Place(
		p.width,
		p.height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}

// SetFocus sets whether this screen has focus
func (p *ProcessingScreen) SetFocus(hasFocus bool) {
	// Processing screen doesn't need special focus handling yet
}