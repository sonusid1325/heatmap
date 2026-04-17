package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"sonusid.in/heatmap/internal/auth"
	"sonusid.in/heatmap/internal/github"
)

// Model is the main TUI model
type Model struct {
	state        State
	username     string
	days         []github.Day
	errMsg       string
	heatmapText  string
	width        int
	height       int
	scrollOffset int
	token        string
}

// State represents the current state of the application
type State int

const (
	LoadingState State = iota
	DisplayState
	ErrorState
)

// New creates a new Model with GitHub token
func New(token string) Model {
	return Model{
		state:       LoadingState,
		username:    "",
		days:        []github.Day{},
		errMsg:      "",
		width:       80,
		height:      24,
		scrollOffset: 0,
		token:       token,
	}
}

// Init initializes the model and loads the authenticated user's data
func (m Model) Init() tea.Cmd {
	token := m.token
	return func() tea.Msg {
		user, err := auth.GetAuthenticatedUser()
		if err != nil {
			return GitHubDataMsg{Error: err}
		}
		days, err := github.Fetch(user, token)
		return GitHubDataMsg{Username: user, Days: days, Error: err}
	}
}

// Update handles updates
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyMsg(msg)
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case GitHubDataMsg:
		if msg.Error != nil {
			m.state = ErrorState
			m.errMsg = msg.Error.Error()
		} else {
			m.username = msg.Username
			m.days = msg.Days
			m.state = DisplayState
			renderer := NewHeatmapRenderer(m.days)
			m.heatmapText = renderer.Render()
		}
		return m, nil
	}
	return m, nil
}

// FetchGitHubCmd fetches GitHub data with token
func FetchGitHubCmd(username string, token string) tea.Cmd {
	return func() tea.Msg {
		days, err := github.Fetch(username, token)
		return GitHubDataMsg{Days: days, Error: err}
	}
}

// GitHubDataMsg is the message for GitHub data
type GitHubDataMsg struct {
	Username string
	Days     []github.Day
	Error    error
}

func (m Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "up", "k":
		if m.state == DisplayState && m.scrollOffset > 0 {
			m.scrollOffset--
		}
	case "down", "j":
		if m.state == DisplayState {
			m.scrollOffset++
		}
	}
	return m, nil
}

// View renders the UI
func (m Model) View() string {
	switch m.state {
	case LoadingState:
		return m.renderLoading()
	case DisplayState:
		return m.renderDisplay()
	case ErrorState:
		return m.renderError()
	default:
		return ""
	}
}

func (m Model) renderLoading() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("86"))

	loadingStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("214")).
		MarginTop(2)

	content := titleStyle.Render("🔥 GitHub Heatmap Viewer") + "\n\n"
	content += loadingStyle.Render("Fetching data for @" + m.username + "...")

	return lipgloss.NewStyle().
		Padding(2).
		Render(content)
}

func (m Model) renderDisplay() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("34")).
		MarginBottom(1)

	heatmapStyle := lipgloss.NewStyle().
		Padding(1)

	statsStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("243")).
		MarginTop(1)

	usernameStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("34")).
		Bold(true)

	content := titleStyle.Render("🔥 GitHub Heatmap") + "\n"
	content += usernameStyle.Render("@" + m.username) + "\n\n"
	content += heatmapStyle.Render(m.heatmapText) + "\n"
	content += statsStyle.Render("Contributions: " + fmt.Sprintf("%d", getTotalContributions(m.days)))
	content += "\n" + statsStyle.Render("(q to quit)")

	return lipgloss.NewStyle().
		Padding(1).
		Render(content)
}

func (m Model) renderError() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("34")).
		MarginBottom(1)

	errorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("196")).
		Padding(1).
		MarginTop(2)

	instructionsStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("243")).
		MarginTop(2)

	content := titleStyle.Render("🔥 GitHub Heatmap") + "\n"
	content += errorStyle.Render("Error: " + m.errMsg) + "\n"
	content += instructionsStyle.Render("(q to quit)")

	return lipgloss.NewStyle().
		Padding(2).
		Render(content)
}

func getTotalContributions(days []github.Day) int {
	total := 0
	for _, day := range days {
		total += day.Count
	}
	return total
}
