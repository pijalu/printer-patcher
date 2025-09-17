package config

// Step represents a single step in a patching action
type Step struct {
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
	Script      string `yaml:"script"`
	Expected    string `yaml:"expected,omitempty"`
}

// Action represents a patching action that can be executed on the printer
type Action struct {
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
	Steps       []Step `yaml:"steps"`
}

// Config represents the configuration file structure
type Config struct {
	Username string   `yaml:"username"`
	Password string   `yaml:"password"`
	Actions  []Action `yaml:"actions"`
}
