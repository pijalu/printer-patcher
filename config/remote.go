package config

import (
	"gopkg.in/yaml.v2"
)

// RemoteConfigLoader handles loading configuration from remote sources
type RemoteConfigLoader struct {
	branch   string
	username string
	password string
	actions  []Action
}

// NewRemoteConfigLoader creates a new remote config loader
func NewRemoteConfigLoader(branch, username, password string) *RemoteConfigLoader {
	return &RemoteConfigLoader{
		branch:   branch,
		username: username,
		password: password,
	}
}

// LoadConfig loads actions from remote YAML configuration
func (r *RemoteConfigLoader) LoadConfig(data []byte) error {
	var config Config
	err := yaml.Unmarshal(data, &config)
	if err != nil {
		return err
	}

	r.username = config.Username
	r.password = config.Password
	r.actions = config.Actions
	return nil
}

// GetUsername returns the SSH username
func (r *RemoteConfigLoader) GetUsername() string {
	return r.username
}

// GetPassword returns the SSH password
func (r *RemoteConfigLoader) GetPassword() string {
	return r.password
}

// GetActions returns the actions
func (r *RemoteConfigLoader) GetActions() []Action {
	return r.actions
}

// LoadStep loads a step script from remote source
func (r *RemoteConfigLoader) LoadStep(step string, data []byte) (string, error) {
	return string(data), nil
}