package models

import (
	"Alice088/essentia/internal/tui/components"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// HomeScreen represents the home screen with file input
type HomeScreen struct {
	width      int
	height     int
	input      *components.Input
	status     *components.Status
	footer     *components.Footer
	hasFocus   bool
	errorTimer tea.Cmd
	errorMsg   string
}

// clearErrorMsg is sent to clear the error state after a delay
type clearErrorMsg struct{}

// NewHomeScreen creates a new HomeScreen
func NewHomeScreen() Screen {
	screen := &HomeScreen{
		input:    components.NewInput(),
		status:   components.NewStatus(),
		footer:   components.NewFooter(),
		hasFocus: true,
	}

	// Initialize components
	screen.input.Init()
	screen.input.SetFocused(true)
	screen.input.SetWidth(50)
	screen.footer.GetStatus().SetIdle("Ready to process files")

	return screen
}

// clearErrorCmd creates a command to clear the error after 2 seconds
func (h *HomeScreen) clearErrorCmd() tea.Cmd {
	return tea.Tick(2*time.Second, func(time.Time) tea.Msg {
		return clearErrorMsg{}
	})
}

// Init initializes the home screen
func (h *HomeScreen) Init() tea.Cmd {
	return nil
}

// Update handles messages for the home screen
func (h *HomeScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	if !h.hasFocus {
		return h, nil
	}

	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h.width = msg.Width
		h.height = msg.Height
		// Update component widths based on window size
		// Ensure minimum terminal size 120x30
		h.width = max(h.width, 120)
		h.height = max(h.height, 30)

		// Input width: responsive, between 60 and 120
		inputWidth := max(min(h.width-40, 120), 60)
		h.input.SetWidth(inputWidth)
		h.footer.SetWidth(h.width)

	case clearErrorMsg:
		// Clear the error state after timer expires
		h.input.SetError(false)
			h.errorMsg = ""

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+e":
			path := h.input.GetText()

			if path != "" {
				if ok, errMsg := h.validateFilePath(path); ok {
					h.input.Clear()
					// TODO: Actually load the file
					return h, nil
				} else {
					h.input.SetError(true)
					h.errorMsg = errMsg
					cmds = append(cmds, h.clearErrorCmd())
				}
			} else {
				h.input.SetError(true)
				h.errorMsg = "Please enter a file path"
				cmds = append(cmds, h.clearErrorCmd())
			}
		case "ctrl+c", "q":
			return h, tea.Quit
		}
	}

	// Update input component
	var inputCmd tea.Cmd
	h.input, inputCmd = h.input.Update(msg)
	if inputCmd != nil {
		cmds = append(cmds, inputCmd)
	}

	return h, tea.Batch(cmds...)
}

// View renders the home screen
func (h *HomeScreen) View() string {
	if h.width == 0 || h.height == 0 {
		return "Loading..."
	}

	// Load ASCII art
	bytes, err := os.ReadFile("./static/essentia_art.txt")
	if err != nil {
		panic("failed to load home art: " + err.Error())
	}
	asciiArt := string(bytes)

	// Quote
	quote := "Admire simplicity, not complexity."

	// Styles
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

	// Input section (without label, placeholder is inside input)
	inputContainer := lipgloss.NewStyle().
		Padding(1, 2).
		Align(lipgloss.Center).
		Width(h.width - 4).
		Render(h.input.View())

	// Combine all content
	var content string
	if h.errorMsg != "" {
		// Error message style
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#C53030")). // red color
			Align(lipgloss.Center).
			MarginTop(1)
		errorView := errorStyle.Render(h.errorMsg)
		content = lipgloss.JoinVertical(
			lipgloss.Center,
			asciiStyle.Render(asciiArt),
			quoteStyle.Render(quote),
			inputContainer,
			errorView,
		)
	} else {
		content = lipgloss.JoinVertical(
			lipgloss.Center,
			asciiStyle.Render(asciiArt),
			quoteStyle.Render(quote),
			inputContainer,
		)
	}

	// Place content in the middle (above footer)
	// Reserve space for footer (3 lines)
	footerHeight := 3
	contentHeight := max(h.height-footerHeight, 10)

	placedContent := lipgloss.Place(
		h.width,
		contentHeight,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)

	// Add footer
	footerView := h.footer.View()

	// Combine content and footer
	return lipgloss.JoinVertical(
		lipgloss.Top,
		placedContent,
		footerView,
	)
}

// validateFilePath checks if a file exists and returns error message
func (h *HomeScreen) validateFilePath(path string) (bool, string) {
	if _, err := os.Stat(path); err != nil {
		return false, "File not found or cannot be accessed"
	}
	return true, ""
}

// SetFocus sets whether this screen has focus
func (h *HomeScreen) SetFocus(hasFocus bool) {
	h.hasFocus = hasFocus
	h.input.SetFocused(hasFocus)
}
