package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	DataPath  string `json:"data_path"`
	IndexPath string `json:"index_path"`
}

// LoadConfig loads configuration from various sources
func LoadConfig() (*Config, error) {
	config := &Config{}

	// 1. Check environment variables first (highest priority)
	if dataPath := os.Getenv("OED_DATA_PATH"); dataPath != "" {
		config.DataPath = dataPath
	}
	if indexPath := os.Getenv("OED_INDEX_PATH"); indexPath != "" {
		config.IndexPath = indexPath
	}

	// 2. Check for config file in current directory
	if config.DataPath == "" || config.IndexPath == "" {
		loadFromFile("oed_config.json", config)
	}

	// 3. Check for config file in home directory
	if config.DataPath == "" || config.IndexPath == "" {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			configPath := filepath.Join(homeDir, ".oed_mcp", "config.json")
			loadFromFile(configPath, config)
		}
	}

	// 4. Check for config in /etc
	if config.DataPath == "" || config.IndexPath == "" {
		loadFromFile("/etc/oed_mcp/config.json", config)
	}

	// Validate that we have both paths
	if config.DataPath == "" || config.IndexPath == "" {
		return nil, fmt.Errorf("OED data paths not configured. Please set OED_DATA_PATH and OED_INDEX_PATH environment variables or create a config file")
	}

	// Verify files exist
	if _, err := os.Stat(config.DataPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("OED data file not found at: %s", config.DataPath)
	}
	if _, err := os.Stat(config.IndexPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("OED index file not found at: %s", config.IndexPath)
	}

	return config, nil
}

func loadFromFile(path string, config *Config) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Decode separately so file values only fill paths that are still empty
	// and never override ones already set by a higher-priority source.
	var fromFile Config
	if err := json.NewDecoder(file).Decode(&fromFile); err != nil {
		return err
	}
	if config.DataPath == "" {
		config.DataPath = fromFile.DataPath
	}
	if config.IndexPath == "" {
		config.IndexPath = fromFile.IndexPath
	}
	return nil
}