package ui

import (
	"fmt"
	"sort"
	"time"

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
		state:        LoadingState,
		username:     "",
		days:         []github.Day{},
		errMsg:       "",
		width:        80,
		height:       24,
		scrollOffset: 0,
		token:        token,
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
		Foreground(lipgloss.Color("46"))

	loadingStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("248")).
		MarginTop(2)

	loadingText := "Authenticating with GitHub..."
	if m.username != "" {
		loadingText = "Fetching contributions for @" + m.username + "..."
	}

	content := titleStyle.Render("🔥 GitHub Heatmap Viewer") + "\n\n"
	content += loadingStyle.Render(loadingText)

	return lipgloss.NewStyle().
		Padding(1, 2).
		Render(content)
}

func (m Model) renderDisplay() string {
	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(1, 2)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("46")).
		MarginBottom(1)

	statsStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")).
		MarginTop(1)

	usernameStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("34")).
		Bold(true)

	total := getTotalContributions(m.days)
	dateRange := getContributionRange(m.days)
	currentStreak, longestStreak := getContributionStreaks(m.days)

	content := titleStyle.Render("GitHub Contribution Heatmap") + "\n"
	content += usernameStyle.Render("@"+m.username) + "\n\n"
	content += m.heatmapText + "\n"
	content += statsStyle.Render(
		fmt.Sprintf(
			"Total: %d • Current streak: %d days • Longest streak: %d days • %s",
			total,
			currentStreak,
			longestStreak,
			dateRange,
		),
	)
	content += "\n" + statsStyle.Render("q: quit")

	return lipgloss.NewStyle().
		Padding(1, 2).
		Render(cardStyle.Render(content))
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
	content += errorStyle.Render("Error: "+m.errMsg) + "\n"
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

func getContributionRange(days []github.Day) string {
	if len(days) == 0 {
		return "No date range"
	}

	parsedDates := make([]time.Time, 0, len(days))
	for _, day := range days {
		date, err := time.Parse("2006-01-02", day.Date)
		if err != nil {
			continue
		}
		parsedDates = append(parsedDates, date)
	}
	if len(parsedDates) == 0 {
		return "No date range"
	}

	sort.Slice(parsedDates, func(i, j int) bool {
		return parsedDates[i].Before(parsedDates[j])
	})

	return fmt.Sprintf("%s – %s", parsedDates[0].Format("Jan 2, 2006"), parsedDates[len(parsedDates)-1].Format("Jan 2, 2006"))
}

func getContributionStreaks(days []github.Day) (int, int) {
	typedDays := make([]github.Day, 0, len(days))
	for _, day := range days {
		if _, err := time.Parse("2006-01-02", day.Date); err == nil {
			typedDays = append(typedDays, day)
		}
	}
	if len(typedDays) == 0 {
		return 0, 0
	}

	sort.Slice(typedDays, func(i, j int) bool {
		return typedDays[i].Date < typedDays[j].Date
	})

	longest := 0
	currentRun := 0
	for _, day := range typedDays {
		if day.Count > 0 {
			currentRun++
			if currentRun > longest {
				longest = currentRun
			}
		} else {
			currentRun = 0
		}
	}

	current := 0
	for i := len(typedDays) - 1; i >= 0; i-- {
		if typedDays[i].Count > 0 {
			current++
		} else {
			break
		}
	}

	return current, longest
}
