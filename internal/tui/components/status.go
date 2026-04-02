package components

import (
	"github.com/charmbracelet/lipgloss"
)

// Status represents the pipeline status component
type Status struct {
	state   string
	message string
	width   int
	style   lipgloss.Style
}

// PipelineState represents possible pipeline states
type PipelineState string

const (
	StateIdle      PipelineState = "idle"
	StateWorking   PipelineState = "wrk"
	StateError     PipelineState = "err"
)

// NewStatus creates a new Status component
func NewStatus() *Status {
	return &Status{
		state:   string(StateIdle),
		message: "Ready to process files",
		width:   30,
		style: lipgloss.NewStyle().
			Padding(0, 1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#4A5568")).
			Align(lipgloss.Center),
	}
}

// View renders the status component (compact version for footer)
func (s *Status) View() string {
	if s.width == 0 {
		s.width = 10 // Compact width for footer
	}

	// Determine color based on state
	var color lipgloss.Color
	switch PipelineState(s.state) {
	case StateIdle:
		color = lipgloss.Color("#4A5568") // Gray
	case StateWorking:
		color = lipgloss.Color("#FF6B35") // Orange (Claude Code color)
	case StateError:
		color = lipgloss.Color("#F56565") // Red
	default:
		color = lipgloss.Color("#4A5568") // Gray
	}

	// Create styled state (compact, no message)
	stateStyle := lipgloss.NewStyle().
		Foreground(color).
		Bold(true).
		Padding(0, 1)

	// Just show the state, no message
	return stateStyle.Render(s.state)
}

// SetState sets the pipeline state
func (s *Status) SetState(state PipelineState) {
	s.state = string(state)
}

// SetMessage sets the status message
func (s *Status) SetMessage(message string) {
	s.message = message
}

// SetWidth sets the width of the status component
func (s *Status) SetWidth(width int) {
	s.width = width
}

// GetState returns the current pipeline state
func (s *Status) GetState() PipelineState {
	return PipelineState(s.state)
}

// GetMessage returns the current status message
func (s *Status) GetMessage() string {
	return s.message
}

// SetIdle sets the status to idle with a message
func (s *Status) SetIdle(message string) {
	s.state = string(StateIdle)
	s.message = message
}

// SetWorking sets the status to working (wrk)
func (s *Status) SetWorking() {
	s.state = string(StateWorking)
	s.message = ""
}

// SetError sets the status to error with a message
func (s *Status) SetError(message string) {
	s.state = string(StateError)
	s.message = message
}