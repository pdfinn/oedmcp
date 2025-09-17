package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Save original env vars
	origDataPath := os.Getenv("OED_DATA_PATH")
	origIndexPath := os.Getenv("OED_INDEX_PATH")
	defer func() {
		os.Setenv("OED_DATA_PATH", origDataPath)
		os.Setenv("OED_INDEX_PATH", origIndexPath)
	}()

	t.Run("EnvironmentVariables", func(t *testing.T) {
		// Create temp files to act as data files
		tempDir := t.TempDir()
		dataPath := filepath.Join(tempDir, "oed2")
		indexPath := filepath.Join(tempDir, "oed2index")

		// Create dummy files
		if err := os.WriteFile(dataPath, []byte("data"), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(indexPath, []byte("index"), 0644); err != nil {
			t.Fatal(err)
		}

		// Set environment variables
		os.Setenv("OED_DATA_PATH", dataPath)
		os.Setenv("OED_INDEX_PATH", indexPath)

		cfg, err := LoadConfig()
		if err != nil {
			t.Fatalf("LoadConfig failed: %v", err)
		}

		if cfg.DataPath != dataPath {
			t.Errorf("DataPath = %q, want %q", cfg.DataPath, dataPath)
		}
		if cfg.IndexPath != indexPath {
			t.Errorf("IndexPath = %q, want %q", cfg.IndexPath, indexPath)
		}
	})

	t.Run("ConfigFile", func(t *testing.T) {
		// Clear env vars
		os.Unsetenv("OED_DATA_PATH")
		os.Unsetenv("OED_INDEX_PATH")

		// Create temp directory
		tempDir := t.TempDir()
		dataPath := filepath.Join(tempDir, "oed2")
		indexPath := filepath.Join(tempDir, "oed2index")

		// Create dummy files
		if err := os.WriteFile(dataPath, []byte("data"), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(indexPath, []byte("index"), 0644); err != nil {
			t.Fatal(err)
		}

		// Change to temp directory
		origWd, _ := os.Getwd()
		os.Chdir(tempDir)
		defer os.Chdir(origWd)

		// Create config file
		configData := Config{
			DataPath:  dataPath,
			IndexPath: indexPath,
		}
		data, _ := json.Marshal(configData)
		if err := os.WriteFile("oed_config.json", data, 0644); err != nil {
			t.Fatal(err)
		}

		cfg, err := LoadConfig()
		if err != nil {
			t.Fatalf("LoadConfig failed: %v", err)
		}

		if cfg.DataPath != dataPath {
			t.Errorf("DataPath = %q, want %q", cfg.DataPath, dataPath)
		}
		if cfg.IndexPath != indexPath {
			t.Errorf("IndexPath = %q, want %q", cfg.IndexPath, indexPath)
		}
	})

	t.Run("MissingConfig", func(t *testing.T) {
		// Clear env vars
		os.Unsetenv("OED_DATA_PATH")
		os.Unsetenv("OED_INDEX_PATH")

		// Change to temp directory with no config
		tempDir := t.TempDir()
		origWd, _ := os.Getwd()
		os.Chdir(tempDir)
		defer os.Chdir(origWd)

		_, err := LoadConfig()
		if err == nil {
			t.Error("LoadConfig should fail with missing config")
		}
	})

	t.Run("InvalidDataPath", func(t *testing.T) {
		// Set invalid paths
		os.Setenv("OED_DATA_PATH", "/nonexistent/path/oed2")
		os.Setenv("OED_INDEX_PATH", "/nonexistent/path/oed2index")

		_, err := LoadConfig()
		if err == nil {
			t.Error("LoadConfig should fail with invalid data path")
		}
	})
}