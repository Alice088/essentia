package components

import (
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Input is a text input component
type Input struct {
	text             string
	placeholder      string
	cursorPos        int
	width            int
	focused          bool
	hasError         bool
	maxLength        int
	errorMsg         string
	errorTimer       tea.Cmd
	style            lipgloss.Style
	errorStyle       lipgloss.Style
	cursorStyle      lipgloss.Style
	placeholderStyle lipgloss.Style
	errorTextStyle   lipgloss.Style
}

// clearErrorMsg is sent to clear the error after a delay
type clearErrorMsg struct{}

// pasteMsg is sent when clipboard paste is requested
type pasteMsg struct {
	text string
}

// NewInput creates a new Input component
func NewInput() *Input {
	return &Input{
		text:        "",
		placeholder: "~/path/to/document.pdf",
		cursorPos:   0,
		focused:     false,
		hasError:    false,
		maxLength:   70, // Limit for paste from clipboard
		style: lipgloss.NewStyle().
			Padding(0, 1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#4A5568")),
		errorStyle: lipgloss.NewStyle().
			Padding(0, 1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#C53030")), // Красный, близкий к основной палитре
		cursorStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6B35")),
		placeholderStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#718096")).
			Faint(true),
		errorTextStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#C53030")). // red color
			Align(lipgloss.Center).
			MarginTop(1),
	}
}

// Init initializes the input component
func (i *Input) Init() tea.Cmd {
	return nil
}

// pasteFromClipboardCmd returns a command to read from clipboard
func (i *Input) pasteFromClipboardCmd() tea.Cmd {
	return func() tea.Msg {
		// Try to read from clipboard using xclip (Linux)
		cmd := exec.Command("xclip", "-selection", "clipboard", "-o")
		output, err := cmd.Output()
		if err != nil {
			// Failed to read clipboard
			return pasteMsg{text: ""}
		}

		// Clean up newlines and other whitespace
		text := strings.TrimSpace(string(output))
		return pasteMsg{text: text}
	}
}

// Update handles messages for the input component
func (i *Input) Update(msg tea.Msg) (*Input, tea.Cmd) {
	if !i.focused {
		return i, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left":
			if i.cursorPos > 0 {
				i.cursorPos--
				i.hasError = false
				i.errorMsg = ""
			}
		case "right":
			if i.cursorPos < len(i.text) {
				i.cursorPos++
				i.hasError = false
				i.errorMsg = ""
			}
		case "backspace":
			if i.cursorPos > 0 {
				i.text = i.text[:i.cursorPos-1] + i.text[i.cursorPos:]
				i.cursorPos--
				i.hasError = false
				i.errorMsg = ""
			}
		case "delete":
			if i.cursorPos < len(i.text) {
				i.text = i.text[:i.cursorPos] + i.text[i.cursorPos+1:]
				i.hasError = false
				i.errorMsg = ""
			}
		case "home":
			i.cursorPos = 0
			i.hasError = false
			i.errorMsg = ""
		case "end":
			i.cursorPos = len(i.text)
			i.hasError = false
			i.errorMsg = ""
		case "ctrl+w":
			// Delete word before cursor
			if i.cursorPos > 0 {
				start := i.cursorPos - 1
				for start > 0 && i.text[start-1] != ' ' {
					start--
				}
				i.text = i.text[:start] + i.text[i.cursorPos:]
				i.cursorPos = start
				i.hasError = false
				i.errorMsg = ""
			}
		case "ctrl+v":
			// Handle clipboard paste with 70 character limit
			return i, i.pasteFromClipboardCmd()
		default:
			if len(msg.String()) == 1 && msg.Type == tea.KeyRunes {
				r := msg.Runes[0]
				// Skip control characters
				if r >= 32 && r <= 126 {
					// Check max length (70 characters limit for paste compatibility)
					if len(i.text) < i.maxLength {
						i.text = i.text[:i.cursorPos] + string(r) + i.text[i.cursorPos:]
						i.cursorPos++
						i.hasError = false
						i.errorMsg = ""
					} else {
						// Text too long - set error
						i.hasError = true
						i.errorMsg = "too long"
						// Start timer to clear error after 2 seconds
						i.errorTimer = tea.Tick(2*time.Second, func(time.Time) tea.Msg {
							return clearErrorMsg{}
						})
						return i, i.errorTimer
					}
				}
			}
		}
	case pasteMsg:
		// Handle paste from clipboard
		pastedText := msg.text

		// Apply 70 character limit - truncate if too long
		if len(pastedText) > i.maxLength {
			pastedText = pastedText[:i.maxLength]
			// Set error message for truncated paste
			i.hasError = true
			i.errorMsg = "too long"
			// Start timer to clear error after 2 seconds
			i.errorTimer = tea.Tick(2*time.Second, func(time.Time) tea.Msg {
				return clearErrorMsg{}
			})
			return i, i.errorTimer
		} else {
			i.hasError = false
			i.errorMsg = ""
		}

		// Insert at cursor position
		i.text = i.text[:i.cursorPos] + pastedText + i.text[i.cursorPos:]
		i.cursorPos += len(pastedText)
	case clearErrorMsg:
		// Clear the error state after timer expires
		i.hasError = false
		i.errorMsg = ""
		i.errorTimer = nil
	}

	// Return any pending error timer
	if i.errorTimer != nil {
		return i, i.errorTimer
	}
	return i, nil
}

// View renders the input component
func (i *Input) View() string {
	if i.width == 0 {
		i.width = 40 // default width
	}

	// Determine what text to display
	var displayText string
	var showPlaceholder bool

	if i.text == "" {
		// Show placeholder
		displayText = i.placeholder
		showPlaceholder = true
	} else {
		// Show actual text
		displayText = i.text
		showPlaceholder = false
	}

	// Truncate text if too long for display
	maxDisplayChars := i.width - 4 // account for border and padding
	if len(displayText) > maxDisplayChars {
		// Show end of text with cursor position consideration
		if i.cursorPos > maxDisplayChars-10 {
			start := max(i.cursorPos-maxDisplayChars+10, 0)
			displayText = "…" + displayText[start:]
		} else {
			displayText = displayText[:maxDisplayChars-1] + "…"
		}
	}

	// Build the view
	view := ""
	if showPlaceholder {
		// Render placeholder with placeholder style
		view = i.placeholderStyle.Render(displayText)
		// Add cursor at the beginning if focused
		if i.focused && i.cursorPos == 0 {
			view = i.cursorStyle.Render(" ") + view
		}
	} else {
		// Render actual text with cursor
		for pos, char := range displayText {
			if pos == i.cursorPos {
				view += i.cursorStyle.Render(string(char))
			} else {
				view += string(char)
			}
		}
		// If cursor is at the end
		if i.cursorPos == len(displayText) {
			view += i.cursorStyle.Render(" ")
		}
	}

	// Apply style - check if we should show error state
	var inputView string

	if i.hasError {
		// Show error style with red border
		inputView = i.errorStyle.Width(i.width).Render(view)
		if i.focused {
			inputView = i.errorStyle.
				BorderForeground(lipgloss.Color("#C53030")).
				Width(i.width).
				Render(view)
		}
	} else {
		// Show normal style
		inputView = i.style.Width(i.width).Render(view)
		if i.focused {
			inputView = i.style.
				BorderForeground(lipgloss.Color("#FF6B35")).
				Width(i.width).
				Render(view)
		}
	}

	// If there's an error message, show it below the input
	if i.errorMsg != "" {
		errorView := i.errorTextStyle.Render(i.errorMsg)
		return lipgloss.JoinVertical(
			lipgloss.Center,
			inputView,
			errorView,
		)
	}

	return inputView
}

// SetText sets the input text
func (i *Input) SetText(text string) {
	i.text = text
	i.cursorPos = len(text)
}

// GetText returns the input text
func (i *Input) GetText() string {
	return i.text
}

// SetFocused sets the focus state
func (i *Input) SetFocused(focused bool) {
	i.focused = focused
}

// SetWidth sets the width of the input
func (i *Input) SetWidth(width int) {
	i.width = width
}

// CursorPosition returns the current cursor position
func (i *Input) CursorPosition() int {
	return i.cursorPos
}

// SetCursorPosition sets the cursor position
func (i *Input) SetCursorPosition(pos int) {
	if pos >= 0 && pos <= len(i.text) {
		i.cursorPos = pos
	}
}

// SetError sets the error state of the input
func (i *Input) SetError(hasError bool) {
	i.hasError = hasError
}

// GetError returns the error message if any
func (i *Input) GetError() string {
	return i.errorMsg
}

// Clear clears the input text and resets error state
func (i *Input) Clear() {
	i.text = ""
	i.cursorPos = 0
	i.hasError = false
	i.errorMsg = ""
}
