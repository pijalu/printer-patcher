package config

import (
	"gopkg.in/yaml.v2"
)

// LoadConfig loads actions from a YAML configuration file
func LoadConfig() (*Config, error) {
	data, err := configsFS.ReadFile("actions.yaml")
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func LoadStep(step string) (string, error) {
	data, err := configsFS.ReadFile(step)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// LoadConfigFromData loads configuration from raw data
func LoadConfigFromData(data []byte) (*Config, error) {
	var config Config
	err := yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// LoadStepFromData loads step content from raw data
func LoadStepFromData(data []byte) string {
	return string(data)
}
