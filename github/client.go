package github

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	githubAPIURL = "https://api.github.com"
)

// Branch represents a GitHub branch
type Branch struct {
	Name   string `json:"name"`
	Commit struct {
		SHA string `json:"sha"`
		URL string `json:"url"`
	} `json:"commit"`
}

// Release represents a GitHub release
type Release struct {
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
	Draft   bool   `json:"draft"`
	Prerelease bool `json:"prerelease"`
}

// Client represents a GitHub API client
type Client struct {
	httpClient *http.Client
	owner      string
	name       string
}

// NewClient creates a new GitHub client with default repository
func NewClient() *Client {
	fmt.Println("Creating GitHub client with default repository")
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		owner: RepoOwner,
		name:  RepoName,
	}
}

// NewClientWithRepo creates a new GitHub client with a specific repository
func NewClientWithRepo(owner, name string) *Client {
	fmt.Printf("Creating GitHub client for repository: %s/%s\n", owner, name)
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		owner: owner,
		name:  name,
	}
}

// GetBranches fetches all branches from the repository
func (c *Client) GetBranches() ([]Branch, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/branches", githubAPIURL, c.owner, c.name)
	fmt.Printf("Fetching branches from: %s\n", url)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	// Set user agent as recommended by GitHub API
	req.Header.Set("User-Agent", "printer-patcher-app")
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error: %s - %s", resp.Status, string(body))
	}
	
	var branches []Branch
	if err := json.NewDecoder(resp.Body).Decode(&branches); err != nil {
		return nil, err
	}
	
	fmt.Printf("Found %d branches for %s/%s\n", len(branches), c.owner, c.name)
	return branches, nil
}

// GetReleases fetches all releases from the repository
func (c *Client) GetReleases() ([]Release, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/releases", githubAPIURL, c.owner, c.name)
	fmt.Printf("Fetching releases from: %s\n", url)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	// Set user agent as recommended by GitHub API
	req.Header.Set("User-Agent", "printer-patcher-app")
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error: %s - %s", resp.Status, string(body))
	}
	
	var releases []Release
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, err
	}
	
	fmt.Printf("Found %d releases for %s/%s\n", len(releases), c.owner, c.name)
	return releases, nil
}

// GetBranchNames returns a list of branch names including "main" and all release branches
func (c *Client) GetBranchNames() ([]string, error) {
	fmt.Printf("Fetching branch names for %s/%s\n", c.owner, c.name)
	// Get branches
	branches, err := c.GetBranches()
	if err != nil {
		return nil, fmt.Errorf("failed to get branches: %w", err)
	}
	
	// Get releases
	releases, err := c.GetReleases()
	if err != nil {
		return nil, fmt.Errorf("failed to get releases: %w", err)
	}
	
	// Collect branch names
	var names []string
	names = append(names, "local") // Always include local as an option
	
	// Add main branch if it exists
	for _, branch := range branches {
		if branch.Name == "main" {
			names = append(names, "main")
			break
		}
	}
	
	// Add release branches
	for _, release := range releases {
		if !release.Draft {
			names = append(names, release.TagName)
		}
	}
	
	fmt.Printf("Available sources for %s/%s: %v\n", c.owner, c.name, names)
	return names, nil
}