package main

import (
"fmt"

"sonusid.in/heatmap/internal/auth"
"sonusid.in/heatmap/internal/github"
"sonusid.in/heatmap/internal/ui"
)

func main() {
token, _ := auth.GetGitHubToken()
user, _ := auth.GetAuthenticatedUser()

days, _ := github.Fetch(user, token)
renderer := ui.NewHeatmapRenderer(days)
output := renderer.Render()

fmt.Printf("🔥 GitHub Heatmap\n\n@%s\n\n", user)
fmt.Println(output)

total := 0
for _, d := range days {
total += d.Count
}
fmt.Printf("Contributions: %d\n", total)
}
