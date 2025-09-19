package github

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Downloader handles downloading content from GitHub
type Downloader struct {
	client *Client
	cache  *Cache
}

// NewDownloader creates a new downloader instance with default repository
func NewDownloader() (*Downloader, error) {
	fmt.Println("Creating GitHub downloader with default repository")
	client := NewClient()
	cache, err := NewCache()
	if err != nil {
		return nil, err
	}
	
	return &Downloader{
		client: client,
		cache:  cache,
	}, nil
}

// NewDownloaderWithRepo creates a new downloader instance with a specific repository
func NewDownloaderWithRepo(owner, name string) (*Downloader, error) {
	fmt.Printf("Creating GitHub downloader for repository: %s/%s\n", owner, name)
	client := NewClientWithRepo(owner, name)
	cache, err := NewCache()
	if err != nil {
		return nil, err
	}
	
	return &Downloader{
		client: client,
		cache:  cache,
	}, nil
}

// DownloadFile downloads a file from GitHub with caching
func (d *Downloader) DownloadFile(branch, filePath string) ([]byte, error) {
	// For local source, we don't download anything
	if branch == "local" {
		return nil, fmt.Errorf("cannot download file for local source")
	}
	
	url := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/%s", d.client.owner, d.client.name, branch, filePath)
	fmt.Printf("Downloading file from: %s\n", url)
	
	// Check cache first
	if content, found, err := d.cache.Get(url); err == nil && found {
		fmt.Printf("File found in cache: %s\n", url)
		return content, nil
	}
	
	// Download from GitHub
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download %s: %w", url, err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download %s: %s", url, resp.Status)
	}
	
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	
	// Cache the content
	if err := d.cache.Put(url, content); err != nil {
		// Log error but don't fail the download
		fmt.Printf("Warning: failed to cache content: %v\n", err)
	}
	
	fmt.Printf("File downloaded successfully: %s (%d bytes)\n", url, len(content))
	return content, nil
}

// DownloadConfig downloads the actions.yaml file from a specific branch
func (d *Downloader) DownloadConfig(branch string) ([]byte, error) {
	fmt.Printf("Downloading config from branch: %s\n", branch)
	return d.DownloadFile(branch, "config/actions.yaml")
}

// DownloadScript downloads a script file from a specific branch
func (d *Downloader) DownloadScript(branch, scriptPath string) ([]byte, error) {
	fmt.Printf("Downloading script from branch %s: %s\n", branch, scriptPath)
	// Normalize script path to be relative to config directory
	scriptPath = strings.TrimPrefix(scriptPath, "config/")
	return d.DownloadFile(branch, "config/"+scriptPath)
}