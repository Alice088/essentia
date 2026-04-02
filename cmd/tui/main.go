package main

import (
	"fmt"
	"os"

	"Alice088/essentia/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	app := tui.NewApp()
	p := tea.NewProgram(
		app,
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running TUI: %v\n", err)
		os.Exit(1)
	}
}
