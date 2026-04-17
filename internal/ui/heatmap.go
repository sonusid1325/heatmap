package ui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"sonusid.in/heatmap/internal/github"
)

const (
	cellFillActive = "▣"
	cellFillEmpty  = "□"
	cellGap        = " "
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
	if len(h.days) == 0 {
		return ""
	}

	dates := make([]time.Time, 0, len(h.days))
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

	sort.Slice(dates, func(i, j int) bool {
		return dates[i].Before(dates[j])
	})

	startDate := dates[0]
	endDate := dates[len(dates)-1]
	for startDate.Weekday() != time.Sunday {
		startDate = startDate.AddDate(0, 0, -1)
	}
	for endDate.Weekday() != time.Saturday {
		endDate = endDate.AddDate(0, 0, 1)
	}

	weekStarts := make([]time.Time, 0)
	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 7) {
		weekStarts = append(weekStarts, d)
	}
	if len(weekStarts) == 0 {
		return ""
	}

	var result strings.Builder

	// Month labels (GitHub-style, shown at month boundaries).
	result.WriteString("    ")
	for i, weekStart := range weekStarts {
		result.WriteString(h.monthLabelForWeek(weekStart))
		if i < len(weekStarts)-1 {
			result.WriteString(" ")
		}
	}
	result.WriteString("\n")

	cellWidth := lipgloss.Width(cellFillActive)
	gapWidth := lipgloss.Width(cellGap)
	innerWidth := len(weekStarts)*cellWidth + (len(weekStarts)-1)*gapWidth
	result.WriteString("    ┌")
	result.WriteString(strings.Repeat("─", innerWidth))
	result.WriteString("┐\n")

	rowLabels := []string{"   ", "Mon", "   ", "Wed", "   ", "Fri", "   "}
	for dayOfWeek := 0; dayOfWeek < 7; dayOfWeek++ {
		result.WriteString(fmt.Sprintf("%3s │", rowLabels[dayOfWeek]))

		for i, weekStart := range weekStarts {
			date := weekStart.AddDate(0, 0, dayOfWeek)
			count := dayMap[date.Format("2006-01-02")]
			result.WriteString(h.renderCell(count))
			if i < len(weekStarts)-1 {
				result.WriteString(cellGap)
			}
		}
		result.WriteString("│\n")
	}

	result.WriteString("    └")
	result.WriteString(strings.Repeat("─", innerWidth))
	result.WriteString("┘\n")

	result.WriteString("      Less ")
	for i := 0; i <= 5; i++ {
		result.WriteString(h.renderLegendCell(i))
		if i < 5 {
			result.WriteString(cellGap)
		}
	}
	result.WriteString(" More")

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
	fill := cellFillActive
	if intensity == 0 {
		fill = cellFillEmpty
	}

	return lipgloss.NewStyle().
		Foreground(h.colorForIntensity(intensity)).
		Render(fill)
}

func (h *HeatmapRenderer) renderLegendCell(intensity int) string {
	fill := cellFillActive
	if intensity == 0 {
		fill = cellFillEmpty
	}

	return lipgloss.NewStyle().
		Foreground(h.colorForIntensity(intensity)).
		Render(fill)
}

func (h *HeatmapRenderer) monthLabelForWeek(weekStart time.Time) string {
	for i := 0; i < 7; i++ {
		d := weekStart.AddDate(0, 0, i)
		if d.Day() == 1 {
			return d.Format("Jan")
		}
	}
	return "   "
}

func (h *HeatmapRenderer) colorForIntensity(intensity int) lipgloss.TerminalColor {
	// Adaptive GitHub-like palette (light + dark terminal themes).
	colors := []lipgloss.TerminalColor{
		lipgloss.AdaptiveColor{Light: "#afb8c1", Dark: "#484f58"}, // empty
		lipgloss.AdaptiveColor{Light: "#9be9a8", Dark: "#0e4429"},
		lipgloss.AdaptiveColor{Light: "#40c463", Dark: "#006d32"},
		lipgloss.AdaptiveColor{Light: "#30a14e", Dark: "#26a641"},
		lipgloss.AdaptiveColor{Light: "#216e39", Dark: "#39d353"},
		lipgloss.AdaptiveColor{Light: "#0e4429", Dark: "#56d364"},
	}

	if intensity < 0 {
		return colors[0]
	}
	if intensity >= len(colors) {
		return colors[len(colors)-1]
	}
	return colors[intensity]
}
