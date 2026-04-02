package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"Alice088/essentia/internal/tui/models"
)

// App represents the main TUI application
type App struct {
	screens      []models.Screen
	activeScreen int
	width        int
	height       int
}

// NewApp creates and initializes a new App
func NewApp() *App {
	app := &App{
		screens: []models.Screen{
			models.NewHomeScreen(),
			models.NewProcessingScreen(),
			models.NewMetricsScreen(),
			models.NewHelpScreen(),
		},
		activeScreen: 0,
	}

	return app
}

// Init initializes the app
func (a *App) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the app state
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "right":
			a.activeScreen = (a.activeScreen + 1) % len(a.screens)
			return a, nil
		case "shift+tab", "left":
			a.activeScreen = (a.activeScreen - 1 + len(a.screens)) % len(a.screens)
			return a, nil
		case "enter":
			// Switch to the active screen
			return a, nil
		case "ctrl+c", "q":
			return a, tea.Quit
		}
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
	}

	// Update the active screen
	var cmd tea.Cmd
	a.screens[a.activeScreen], cmd = a.screens[a.activeScreen].Update(msg)
	return a, cmd
}

// View renders the app UI
func (a *App) View() string {
	if a.width == 0 || a.height == 0 {
		return "Initializing..."
	}

	// Header with navigation
	header := a.renderHeader()

	// Active screen content
	content := a.screens[a.activeScreen].View()

	// Footer
	footer := a.renderFooter()

	// Combine everything
	view := lipgloss.JoinVertical(
		lipgloss.Top,
		header,
		content,
		footer,
	)

	return view
}

// renderHeader renders the header with navigation tabs
func (a *App) renderHeader() string {
	tabs := []string{"Home", "Processing", "Metrics", "How does it work"}

	var tabViews []string
	for i, tab := range tabs {
		style := styles.Tab
		if i == a.activeScreen {
			style = styles.ActiveTab
		}
		tabViews = append(tabViews, style.Render(tab))
	}

	header := lipgloss.JoinHorizontal(
		lipgloss.Left,
		tabViews...,
	)

	// Center the header
	return lipgloss.PlaceHorizontal(
		a.width,
		lipgloss.Center,
		header,
	)
}

// renderFooter renders the footer with instructions
func (a *App) renderFooter() string {
	instructions := []string{
		"Tab/→: Next tab",
		"Shift+Tab/←: Prev tab",
		"Enter: Select",
		"Q/Ctrl+C: Quit",
	}

	instructionText := strings.Join(instructions, " • ")

	// Fill the entire width with the footer
	footer := styles.Footer.
		Width(a.width).
		Align(lipgloss.Center).
		Render(instructionText)

	return footer
}