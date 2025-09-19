package main

import (
	"fmt"
	"github.com/pijalu/printer-patcher/config"
	"github.com/pijalu/printer-patcher/github"
)

// ConfigManager handles loading configuration from different sources
type ConfigManager struct {
	localConfig  *config.Config
	remoteConfig *config.Config
	downloader   *github.Downloader
	currentSource string
}

// NewConfigManager creates a new configuration manager
func NewConfigManager(localConfig *config.Config) (*ConfigManager, error) {
	downloader, err := github.NewDownloader()
	if err != nil {
		return nil, fmt.Errorf("failed to create downloader: %w", err)
	}
	
	return &ConfigManager{
		localConfig:  localConfig,
		downloader:   downloader,
		currentSource: "local",
	}, nil
}

// LoadSource loads configuration from the specified source
func (cm *ConfigManager) LoadSource(source string) error {
	if source == "local" {
		cm.currentSource = source
		cm.remoteConfig = nil
		return nil
	}
	
	// Download remote configuration
	configData, err := cm.downloader.DownloadConfig(source)
	if err != nil {
		return fmt.Errorf("failed to download config from %s: %w", source, err)
	}
	
	remoteConfig, err := config.LoadConfigFromData(configData)
	if err != nil {
		return fmt.Errorf("failed to parse remote config: %w", err)
	}
	
	cm.currentSource = source
	cm.remoteConfig = remoteConfig
	return nil
}

// GetCurrentConfig returns the current configuration based on the selected source
func (cm *ConfigManager) GetCurrentConfig() *config.Config {
	if cm.currentSource == "local" || cm.remoteConfig == nil {
		return cm.localConfig
	}
	return cm.remoteConfig
}

// LoadStep loads a step script, handling both local and remote sources
func (cm *ConfigManager) LoadStep(scriptPath string) (string, error) {
	if cm.currentSource == "local" || cm.remoteConfig == nil {
		// Load from local embedded files
		return config.LoadStep(scriptPath)
	}
	
	// Load from remote source
	scriptData, err := cm.downloader.DownloadScript(cm.currentSource, scriptPath)
	if err != nil {
		return "", fmt.Errorf("failed to download script %s from %s: %w", scriptPath, cm.currentSource, err)
	}
	
	return config.LoadStepFromData(scriptData), nil
}