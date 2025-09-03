package main

import (
	"time"

	"gopkg.in/yaml.v2"
	"os"
)

// Config holds the configuration for Git Air
type Config struct {
	WatchInterval time.Duration `yaml:"watch_interval"`
	PullInterval  time.Duration `yaml:"pull_interval"`
	AutoCommit    bool          `yaml:"auto_commit"`
	AutoPush      bool          `yaml:"auto_push"`
	AutoPull      bool          `yaml:"auto_pull"`
	CommitMessage string        `yaml:"commit_message"`
	ExcludePaths  []string      `yaml:"exclude_paths"`
	LogLevel      string        `yaml:"log_level"`
	// Multi-repo settings
	MultiRepo     bool          `yaml:"multi_repo"`
	ScanPaths     []string      `yaml:"scan_paths"`
	MaxRepos      int           `yaml:"max_repos"`
	ScanInterval  time.Duration `yaml:"scan_interval"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		WatchInterval: 30 * time.Second,
		PullInterval:  60 * time.Second,
		AutoCommit:    true,
		AutoPush:      true,
		AutoPull:      true,
		CommitMessage: "auto commit",
		ExcludePaths: []string{
			"node_modules",
			"*.log",
			"*.tmp",
			".DS_Store",
			"vendor",
			"target",
			"build",
		},
		LogLevel: "info",
		// Multi-repo defaults - ENABLED BY DEFAULT for dev server use
		MultiRepo:    true,  // Default to multi-repo mode
		ScanPaths:    []string{"."},
		MaxRepos:     100,  // Increased for dev servers
		ScanInterval: 5 * time.Minute,
	}
}

// LoadConfig loads configuration from a YAML file
func LoadConfig(path string) (*Config, error) {
	config := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		return config, err
	}

	err = yaml.Unmarshal(data, config)
	return config, err
}

// SaveConfig saves configuration to a YAML file
func (c *Config) SaveConfig(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}