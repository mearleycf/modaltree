package main

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	configDir  = ".config/modaltree"
	configFile = "config.yaml"
)

// LoadConfig loads configuration from file or returns defaults
func LoadConfig() (Config, error) {
	config := defaultConfig()
	
	configPath, err := getConfigPath()
	if err != nil {
		return config, err
	}

	data, err := os.ReadFile(configPath)
	if os.IsNotExist(err) {
		// Create default config
		if err := SaveConfig(config); err != nil {
			return config, err
		}
		return config, nil
	} else if err != nil {
		return config, err
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return config, err
	}

	return config, nil
}

// SaveConfig saves the configuration to file
func SaveConfig(config Config) error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	// Ensure config directory exists
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

// defaultConfig returns default configuration values
func defaultConfig() Config {
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "/"
	}

	display := DefaultDisplayConfig()
	
	return Config{
		ShowHidden:     true,
		Editor:         "code",
		ConfirmActions: true,
		CurrentDir:     cwd,
		Display:        display,
		icons:          UnicodeIconSet(),
		treeSymbols:    UnicodeTreeSymbols(),
	}
}

// getConfigPath returns the full path to config file
func getConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, configDir, configFile), nil
}