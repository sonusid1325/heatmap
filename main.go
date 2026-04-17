package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	counter int
}

func initialModel() model {
	return model{
		counter: 0,
	}
}

// Init runs once
func (m model) Init() tea.Cmd {
	return nil
}

// Update handles events
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "up":
			m.counter++

		case "down":
			m.counter--
		}
	}

	return m, nil
}

// View renders UI
func (m model) View() string {
	return fmt.Sprintf(
		"Heatmap CLI 🔥\n\nCounter: %d\n\n↑/↓ to change • q to quit\n",
		m.counter,
	)
}

func main() {
	p := tea.NewProgram(initialModel())

	if _, err := p.Run(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}

