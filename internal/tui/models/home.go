package models

import (
	"os"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// HomeScreen represents the home screen
type HomeScreen struct {
	width  int
	height int
}

// NewHomeScreen creates a new HomeScreen
func NewHomeScreen() Screen {
	return &HomeScreen{}
}

// Init initializes the home screen
func (h *HomeScreen) Init() tea.Cmd {
	return nil
}

// Update handles messages for the home screen
func (h *HomeScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h.width = msg.Width
		h.height = msg.Height
	}
	return h, nil
}

// View renders the home screen
func (h *HomeScreen) View() string {
	if h.width == 0 || h.height == 0 {
		return "Loading..."
	}

	bytes, err := os.ReadFile("./static/essentia_art.txt")
	if err != nil {
		panic("failed to load home art: " + err.Error())
	}
	// ASCII art for "essentia"
	asciiArt := string(bytes)

	// Quote
	quote := "Admire simplicity, not complexity."

	asciiStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF6B35")).
		Align(lipgloss.Center).
		MarginBottom(1)

	quoteStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#4A5568")).
		Italic(true).
		Align(lipgloss.Center).
		MarginTop(1).
		MarginBottom(2)

	contentStyle := lipgloss.NewStyle().
		Padding(1, 2).
		Align(lipgloss.Center).
		Width(h.width - 4)

	message := "Welcome to the Essentia TUI\n\n" +
		"Use Tab/Shift+Tab to navigate between screens\n" +
		"Press Enter to select the active tab\n" +
		"Press Q or Ctrl+C to quit"

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		asciiStyle.Render(asciiArt),
		quoteStyle.Render(quote),
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
