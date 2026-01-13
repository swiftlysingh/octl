package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	configDirName  = "octl"
	configFileName = "config.json"
)

// Config holds the application configuration
type Config struct {
	ClientID string `json:"client_id,omitempty"`
}

// configDir returns the configuration directory path
func configDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ".config", configDirName), nil
}

// ConfigDir returns the configuration directory path (exported)
func ConfigDir() (string, error) {
	return configDir()
}

// configPath returns the full path to the config file
func configPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, configFileName), nil
}

// Load reads the configuration from disk
func Load() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &cfg, nil
}

// Save writes the configuration to disk
func Save(cfg *Config) error {
	dir, err := configDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	path, err := configPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// GetClientID returns the client ID from config or environment
func GetClientID() string {
	// Environment variable takes precedence
	if id := os.Getenv("OCTL_CLIENT_ID"); id != "" {
		return id
	}

	// Try loading from config file
	cfg, err := Load()
	if err != nil {
		return ""
	}

	return cfg.ClientID
}

// SetClientID saves the client ID to the config file
func SetClientID(clientID string) error {
	cfg, err := Load()
	if err != nil {
		cfg = &Config{}
	}
	cfg.ClientID = clientID
	return Save(cfg)
}
