package components

import (
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
	style            lipgloss.Style
	errorStyle       lipgloss.Style
	cursorStyle      lipgloss.Style
	placeholderStyle lipgloss.Style
}

// NewInput creates a new Input component
func NewInput() *Input {
	return &Input{
		text:        "",
		placeholder: "~/path/to/document.pdf",
		cursorPos:   0,
		focused:     false,
		hasError:    false,
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
	}
}

// Init initializes the input component
func (i *Input) Init() tea.Cmd {
	return nil
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
			}
		case "right":
			if i.cursorPos < len(i.text) {
				i.cursorPos++
				i.hasError = false
			}
		case "backspace":
			if i.cursorPos > 0 {
				i.text = i.text[:i.cursorPos-1] + i.text[i.cursorPos:]
				i.cursorPos--
				i.hasError = false
			}
		case "delete":
			if i.cursorPos < len(i.text) {
				i.text = i.text[:i.cursorPos] + i.text[i.cursorPos+1:]
				i.hasError = false
			}
		case "home":
			i.cursorPos = 0
			i.hasError = false
		case "end":
			i.cursorPos = len(i.text)
			i.hasError = false
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
			}
		default:
			if len(msg.String()) == 1 && msg.Type == tea.KeyRunes {
				r := msg.Runes[0]
				// Skip control characters
				if r >= 32 && r <= 126 {
					i.text = i.text[:i.cursorPos] + string(r) + i.text[i.cursorPos:]
					i.cursorPos++
					i.hasError = false
				}
			}
		}
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
	var result string

	if i.hasError {
		// Show error style with red border
		result = i.errorStyle.Width(i.width).Render(view)
		if i.focused {
			result = i.errorStyle.
				BorderForeground(lipgloss.Color("#C53030")).
				Width(i.width).
				Render(view)
		}
	} else {
		// Show normal style
		result = i.style.Width(i.width).Render(view)
		if i.focused {
			result = i.style.
				BorderForeground(lipgloss.Color("#FF6B35")).
				Width(i.width).
				Render(view)
		}
	}

	return result
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

// Clear clears the input text and resets error state
func (i *Input) Clear() {
	i.text = ""
	i.cursorPos = 0
	i.hasError = false
}
