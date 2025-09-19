package config

import (
	"embed"
	"fmt"
	"gopkg.in/yaml.v2"
)

// Repo represents a single repository
type Repo struct {
	Owner string `yaml:"owner"`
	Name  string `yaml:"name"`
}

// RepoConfig represents the repository configuration
type RepoConfig struct {
	Repositories []Repo `yaml:"repositories"`
}

//go:embed repo.yaml
var repoConfigFS embed.FS

// GetRepoConfig returns the repository configuration
func GetRepoConfig() (*RepoConfig, error) {
	data, err := repoConfigFS.ReadFile("repo.yaml")
	if err != nil {
		return nil, err
	}

	var config RepoConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	// If no repositories are defined, add default
	if len(config.Repositories) == 0 {
		config.Repositories = append(config.Repositories, Repo{
			Owner: "pijalu",
			Name:  "printer-patcher",
		})
	}

	return &config, nil
}

// GetDefaultRepo returns the first repository in the list (default)
func (rc *RepoConfig) GetDefaultRepo() *Repo {
	if len(rc.Repositories) > 0 {
		return &rc.Repositories[0]
	}
	return nil
}

// GetRepoIdentifier returns a unique identifier for a repository (owner/name)
func (r *Repo) GetRepoIdentifier() string {
	return fmt.Sprintf("%s/%s", r.Owner, r.Name)
}