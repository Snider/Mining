package mining

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// GetVersion returns the version of the application
func GetVersion() string {
	return version
}

// GetCommit returns the git commit hash
func GetCommit() string {
	return commit
}

// GetBuildDate returns the build date
func GetBuildDate() string {
	return date
}

// GitHubRelease represents the structure of a GitHub release response.
type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
}

// FetchLatestGitHubVersion fetches the latest release version from a GitHub repository.
// It takes the repository owner and name (e.g., "xmrig", "xmrig") and returns the tag name.
// Uses a circuit breaker to prevent cascading failures when GitHub API is unavailable.
func FetchLatestGitHubVersion(owner, repo string) (string, error) {
	cb := getGitHubCircuitBreaker()

	result, err := cb.Execute(func() (interface{}, error) {
		return fetchGitHubVersionDirect(owner, repo)
	})

	if err != nil {
		// If circuit is open, try to return cached value with warning
		if err == ErrCircuitOpen {
			if cached, ok := cb.GetCached(); ok {
				if tagName, ok := cached.(string); ok {
					return tagName, nil
				}
			}
			return "", fmt.Errorf("github API unavailable (circuit breaker open): %w", err)
		}
		return "", err
	}

	tagName, ok := result.(string)
	if !ok {
		return "", fmt.Errorf("unexpected result type from circuit breaker")
	}

	return tagName, nil
}

// fetchGitHubVersionDirect is the actual GitHub API call, wrapped by circuit breaker
func fetchGitHubVersionDirect(owner, repo string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)

	resp, err := getHTTPClient().Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch version: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		io.Copy(io.Discard, resp.Body) // Drain body to allow connection reuse
		return "", fmt.Errorf("failed to get latest release: unexpected status code %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", fmt.Errorf("failed to decode release: %w", err)
	}

	return release.TagName, nil
}
