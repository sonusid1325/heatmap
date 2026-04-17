package auth

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GetGitHubToken retrieves the GitHub token from gh CLI
func GetGitHubToken() (string, error) {
	// Try using gh CLI to get token
	cmd := exec.Command("gh", "auth", "token")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get GitHub token from gh CLI: %w", err)
	}

	token := strings.TrimSpace(string(output))
	if token == "" {
		return "", fmt.Errorf("no token found")
	}

	return token, nil
}

// GetGitHubHost returns the GitHub host from gh config
func GetGitHubHost() (string, error) {
	// Default to github.com
	cmd := exec.Command("gh", "config", "get", "host")
	output, err := cmd.Output()
	if err != nil {
		return "github.com", nil
	}

	host := strings.TrimSpace(string(output))
	if host == "" {
		return "github.com", nil
	}

	return host, nil
}

// GetAuthenticatedUser returns the currently authenticated GitHub user
func GetAuthenticatedUser() (string, error) {
	cmd := exec.Command("gh", "api", "user", "-q", ".login")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get authenticated user: %w", err)
	}

	user := strings.TrimSpace(string(output))
	if user == "" {
		return "", fmt.Errorf("no user found")
	}

	return user, nil
}

// IsGhCliInstalled checks if gh CLI is installed
func IsGhCliInstalled() bool {
	_, err := exec.LookPath("gh")
	return err == nil
}

// GetGhConfigDir returns the gh config directory
func GetGhConfigDir() string {
	if dir, exists := os.LookupEnv("GH_CONFIG_DIR"); exists {
		return dir
	}

	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, ".config", "gh")
	}

	return ""
}
