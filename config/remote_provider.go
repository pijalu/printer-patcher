package config

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"github.com/pijalu/printer-patcher/github"
)

// RemoteConfigProvider provides configuration and scripts from remote GitHub sources
type RemoteConfigProvider struct {
	branch     string
	owner      string
	name       string
	downloader *github.Downloader
}

// NewRemoteConfigProvider creates a new remote config provider with default repository
func NewRemoteConfigProvider(branch string) (*RemoteConfigProvider, error) {
	fmt.Printf("Creating RemoteConfigProvider for branch: %s (default repository)\n", branch)
	
	// Get default repository
	repoConfig, err := GetRepoConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get repo config: %w", err)
	}
	
	defaultRepo := repoConfig.GetDefaultRepo()
	if defaultRepo == nil {
		return nil, fmt.Errorf("no default repository configured")
	}
	
	downloader, err := github.NewDownloaderWithRepo(defaultRepo.Owner, defaultRepo.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to create downloader: %w", err)
	}
	
	return &RemoteConfigProvider{
		branch:     branch,
		owner:      defaultRepo.Owner,
		name:       defaultRepo.Name,
		downloader: downloader,
	}, nil
}

// NewRemoteConfigProviderWithRepo creates a new remote config provider with a specific repository
func NewRemoteConfigProviderWithRepo(branch, owner, name string) (*RemoteConfigProvider, error) {
	fmt.Printf("Creating RemoteConfigProvider for branch: %s (repository: %s/%s)\n", branch, owner, name)
	
	downloader, err := github.NewDownloaderWithRepo(owner, name)
	if err != nil {
		return nil, fmt.Errorf("failed to create downloader: %w", err)
	}
	
	return &RemoteConfigProvider{
		branch:     branch,
		owner:      owner,
		name:       name,
		downloader: downloader,
	}, nil
}

// LoadConfig loads configuration from remote GitHub source
func (r *RemoteConfigProvider) LoadConfig() (*Config, error) {
	fmt.Printf("Loading config from remote source: %s (repository: %s/%s)\n", r.branch, r.owner, r.name)
	configData, err := r.downloader.DownloadConfig(r.branch)
	if err != nil {
		return nil, fmt.Errorf("failed to download config from %s: %w", r.branch, err)
	}
	
	var config Config
	err = yaml.Unmarshal(configData, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse remote config: %w", err)
	}
	
	fmt.Printf("Loaded %d actions from remote source: %s (repository: %s/%s)\n", len(config.Actions), r.branch, r.owner, r.name)
	return &config, nil
}

// LoadStep loads a step script from remote GitHub source
func (r *RemoteConfigProvider) LoadStep(scriptPath string) (string, error) {
	fmt.Printf("Loading step script from remote source %s: %s (repository: %s/%s)\n", r.branch, scriptPath, r.owner, r.name)
	scriptData, err := r.downloader.DownloadScript(r.branch, scriptPath)
	if err != nil {
		return "", fmt.Errorf("failed to download script %s from %s: %w", scriptPath, r.branch, err)
	}
	
	return string(scriptData), nil
}

// GetSourceName returns the name of the source
func (r *RemoteConfigProvider) GetSourceName() string {
	return r.branch
}

// GetRepoIdentifier returns the repository identifier (owner/name)
func (r *RemoteConfigProvider) GetRepoIdentifier() string {
	return fmt.Sprintf("%s/%s", r.owner, r.name)
}