package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"sonusid.in/heatmap/internal/github"
)

// HeatmapRenderer renders GitHub contribution heatmap
type HeatmapRenderer struct {
	days []github.Day
}

// NewHeatmapRenderer creates a new heatmap renderer
func NewHeatmapRenderer(days []github.Day) *HeatmapRenderer {
	return &HeatmapRenderer{days: days}
}

// Render returns the heatmap as a string (GitHub-style)
func (h *HeatmapRenderer) Render() string {
	if len(h.days) == 0 {
		return "No contribution data found"
	}

	// Create a map of dates to contributions
	dayMap := make(map[string]int)
	for _, day := range h.days {
		dayMap[day.Date] = day.Count
	}

	return h.renderGitHubStyle(dayMap)
}

func (h *HeatmapRenderer) renderGitHubStyle(dayMap map[string]int) string {
	var result strings.Builder

	// Get date range
	if len(h.days) == 0 {
		return ""
	}

	// Parse dates
	var dates []time.Time
	for _, day := range h.days {
		t, err := time.Parse("2006-01-02", day.Date)
		if err != nil {
			continue
		}
		dates = append(dates, t)
	}

	if len(dates) == 0 {
		return ""
	}

	// Sort dates
	for i := 0; i < len(dates); i++ {
		for j := i + 1; j < len(dates); j++ {
			if dates[j].Before(dates[i]) {
				dates[i], dates[j] = dates[j], dates[i]
			}
		}
	}

	startDate := dates[0]
	endDate := dates[len(dates)-1]

	// Find Sunday of the week containing startDate
	for startDate.Weekday() != time.Sunday {
		startDate = startDate.AddDate(0, 0, -1)
	}

	// Render day names row
	dayNames := []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
	result.WriteString("     ")
	for _, name := range dayNames {
		result.WriteString(name + "  ")
	}
	result.WriteString("\n")

	// Render the heatmap grid (7 rows for days of week)
	for dayOfWeek := 0; dayOfWeek < 7; dayOfWeek++ {
		dayName := dayNames[dayOfWeek]
		result.WriteString(fmt.Sprintf("%s    ", dayName[:1]))

		// Render columns (weeks)
		for d := startDate; d.Before(endDate) || d.Equal(endDate); d = d.AddDate(0, 0, 1) {
			if int(d.Weekday()) == dayOfWeek {
				count := dayMap[d.Format("2006-01-02")]
				cell := h.renderCell(count)
				result.WriteString(cell + " ")
			}
		}
		result.WriteString("\n")
	}

	return result.String()
}

func (h *HeatmapRenderer) getIntensity(count int) int {
	if count == 0 {
		return 0
	}
	if count < 5 {
		return 1
	}
	if count < 10 {
		return 2
	}
	if count < 20 {
		return 3
	}
	if count < 30 {
		return 4
	}
	return 5
}

func (h *HeatmapRenderer) renderCell(count int) string {
	intensity := h.getIntensity(count)
	// GitHub green colors (from dark to bright)
	colors := []lipgloss.Color{
		lipgloss.Color("237"),  // #0e1117 (very dark/empty)
		lipgloss.Color("22"),   // dark green
		lipgloss.Color("28"),   // medium dark green
		lipgloss.Color("34"),   // medium green
		lipgloss.Color("40"),   // bright green
		lipgloss.Color("46"),   // brightest green
	}

	style := lipgloss.NewStyle().
		Background(colors[intensity]).
		Foreground(lipgloss.Color("0")).
		Width(2).
		Height(1).
		Align(lipgloss.Center)

	return style.Render("█")
}

func (h *HeatmapRenderer) renderEmptyCell() string {
	style := lipgloss.NewStyle().
		Background(lipgloss.Color("237")).
		Foreground(lipgloss.Color("0"))

	return style.Render(" ")
}
