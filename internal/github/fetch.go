package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Day struct {
	Date  string
	Count int
}

type ContributionDay struct {
	Date                string `json:"date"`
	ContributionCount   int    `json:"contributionCount"`
	ContributionLevel   string `json:"contributionLevel"`
}

type ContributionWeek struct {
	ContributionDays []ContributionDay `json:"contributionDays"`
}

type ContributionCollection struct {
	TotalContributions int                 `json:"totalContributions"`
	ContributionCalendar map[string]interface{} `json:"contributionCalendar"`
}

type GraphQLResponse struct {
	Data struct {
		User struct {
			ContributionsCollection struct {
				ContributionCalendar struct {
					TotalContributions int
					Weeks              []ContributionWeek
				}
			}
		}
	}
	Errors []map[string]interface{}
}

const graphqlQuery = `
query($userName:String!) {
  user(login: $userName) {
    contributionsCollection {
      contributionCalendar {
        totalContributions
        weeks {
          contributionDays {
            contributionCount
            date
            contributionLevel
          }
        }
      }
    }
  }
}
`

// Fetch retrieves contribution data for a GitHub user using authenticated API
func Fetch(username string, token string) ([]Day, error) {
	query := map[string]interface{}{
		"query": graphqlQuery,
		"variables": map[string]interface{}{
			"userName": username,
		},
	}

	payload, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("POST", "https://api.github.com/graphql", bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("User-Agent", "GitHub-Heatmap-Viewer")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var graphResp GraphQLResponse
	err = json.Unmarshal(body, &graphResp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(graphResp.Errors) > 0 {
		errMsg := fmt.Sprintf("%v", graphResp.Errors[0])
		return nil, fmt.Errorf("GraphQL error: %s", errMsg)
	}

	var days []Day

	weeks := graphResp.Data.User.ContributionsCollection.ContributionCalendar.Weeks

	for _, week := range weeks {
		for _, day := range week.ContributionDays {
			days = append(days, Day{
				Date:  day.Date,
				Count: day.ContributionCount,
			})
		}
	}

	if len(days) == 0 {
		return nil, fmt.Errorf("no contribution data found for user %s", username)
	}

	return days, nil
}
