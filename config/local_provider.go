package config

import (
	"fmt"
	"gopkg.in/yaml.v2"
)

// LocalConfigProvider provides configuration and scripts from local embedded files
type LocalConfigProvider struct{}

// NewLocalConfigProvider creates a new local config provider
func NewLocalConfigProvider() *LocalConfigProvider {
	fmt.Println("Creating LocalConfigProvider")
	return &LocalConfigProvider{}
}

// LoadConfig loads configuration from local embedded files
func (l *LocalConfigProvider) LoadConfig() (*Config, error) {
	fmt.Println("Loading config from local embedded files")
	data, err := configsFS.ReadFile("actions.yaml")
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Loaded %d actions from local source\n", len(config.Actions))
	return &config, nil
}

// LoadStep loads a step script from local embedded files
func (l *LocalConfigProvider) LoadStep(step string) (string, error) {
	fmt.Printf("Loading step script from local: %s\n", step)
	data, err := configsFS.ReadFile(step)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// GetSourceName returns the name of the source
func (l *LocalConfigProvider) GetSourceName() string {
	return "local"
}