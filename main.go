package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"sonusid.in/heatmap/internal/auth"
	"sonusid.in/heatmap/internal/ui"
)

func main() {
	// Check if user is authenticated
	token, err := auth.GetGitHubToken()
	if err != nil {
		fmt.Fprintln(os.Stderr, "GitHub authentication required!")
		fmt.Fprintln(os.Stderr, "Run: gh auth login")
		os.Exit(1)
	}

	m := ui.New(token)
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
