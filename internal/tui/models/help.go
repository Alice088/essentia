package models

import (
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// HelpScreen represents the "How does it work" screen
type HelpScreen struct {
	width  int
	height int
}

// NewHelpScreen creates a new HelpScreen
func NewHelpScreen() Screen {
	return &HelpScreen{}
}

// Init initializes the help screen
func (h *HelpScreen) Init() tea.Cmd {
	return nil
}

// Update handles messages for the help screen
func (h *HelpScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h.width = msg.Width
		h.height = msg.Height
	}
	return h, nil
}

// View renders the help screen
func (h *HelpScreen) View() string {
	if h.width == 0 || h.height == 0 {
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

	title := titleStyle.Render("How Does It Work")
	message := "This screen will show:\n" +
		"• Documentation about how the application works\n" +
		"• Text file reading from docs/how-it-works.txt\n" +
		"• Scrollable text viewer\n" +
		"• Word wrap for long lines\n" +
		"• Scroll progress indicator\n\n" +
		"To be implemented in Phase 5"

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		contentStyle.Render(message),
	)

	return lipgloss.Place(
		h.width,
		h.height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}

// SetFocus sets whether this screen has focus
func (h *HelpScreen) SetFocus(hasFocus bool) {
	// Help screen doesn't need special focus handling yet
}