package config

// ConfigProvider is an interface for loading configuration and scripts from different sources
type ConfigProvider interface {
	// LoadConfig loads the configuration from the source
	LoadConfig() (*Config, error)
	
	// LoadStep loads a step script from the source
	LoadStep(scriptPath string) (string, error)
	
	// GetSourceName returns the name of the source
	GetSourceName() string
}