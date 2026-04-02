package components

import (
	"github.com/charmbracelet/lipgloss"
)

// Footer represents a footer component with instructions and status
type Footer struct {
	instructions      string
	status            *Status
	width             int
	style             lipgloss.Style
	instructionsStyle lipgloss.Style
}

// NewFooter creates a new Footer component
func NewFooter() *Footer {
	return &Footer{
		instructions: "Tab/→ Next • Shift+Tab/← Prev • Ctrl+E Confirm • Q/Ctrl+C Quit",
		status:       NewStatus(),
		width:        0,
		style: lipgloss.NewStyle().
			Padding(1, 1).
			Border(lipgloss.NormalBorder(), true, false, false, false).
			BorderForeground(lipgloss.Color("#2D3748")).
			Foreground(lipgloss.Color("#4A5568")).
			Width(0),
		instructionsStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#718096")),
	}
}

// View renders the footer component
func (f *Footer) View() string {
	if f.width == 0 {
		f.width = 80 // default width
	}

	// Set status width (compact)
	f.status.SetWidth(10)

	// Create left part: instructions
	leftPart := f.instructionsStyle.Render(f.instructions)

	// Create right part: status
	rightPart := f.status.View()

	// Calculate widths
	statusWidth := 10                              // fixed width for status
	instructionsWidth := f.width - statusWidth - 2 // minus separator spaces

	// Truncate instructions if too long
	instructions := f.instructions
	if len(instructions) > instructionsWidth {
		instructions = instructions[:instructionsWidth-3] + "..."
	}
	leftPart = f.instructionsStyle.Width(instructionsWidth).Render(instructions)

	// Combine left and right parts
	content := lipgloss.JoinHorizontal(
		lipgloss.Left,
		leftPart,
		lipgloss.NewStyle().Width(f.width-instructionsWidth-statusWidth).Render(""),
		rightPart,
	)

	// Apply container style
	return f.style.Width(f.width).Render(content)
}

// SetWidth sets the width of the footer
func (f *Footer) SetWidth(width int) {
	f.width = width
}

// GetStatus returns the status component
func (f *Footer) GetStatus() *Status {
	return f.status
}

// SetInstructions sets the instructions text
func (f *Footer) SetInstructions(instructions string) {
	f.instructions = instructions
}
