package config

import (
	"fmt"
	"os"
	"path/filepath"

	"go.yaml.in/yaml/v3"
)

// loadFirstYAMLConfig loads the first config file found in CONFIG_FILE or
// the default search paths. It reports whether a file was loaded.
func loadFirstYAMLConfig(config *Config) (bool, error) {
	yamlLoaded := false
	configPaths := []string{
		"config.yaml",                                // Current directory
		"config.yml",                                 // Current directory, alternative extension
		filepath.Join("config", "config.yaml"),       // Config subdirectory
		filepath.Join("config", "config.yml"),        // Config subdirectory, alternative extension
		filepath.Join("..", "config.yaml"),           // Parent directory
		filepath.Join("..", "config.yml"),            // Parent directory, alternative extension
		filepath.Join("..", "config", "config.yaml"), // Parent's config subdirectory
		filepath.Join("..", "config", "config.yml"),  // Parent's config subdirectory, alternative extension
	}

	// Also check if CONFIG_FILE env var is set
	if configFile := os.Getenv("CONFIG_FILE"); configFile != "" {
		configPaths = append([]string{configFile}, configPaths...)
	}

	// Try each path
	for _, path := range configPaths {
		err := loadYAMLConfig(config, path)
		if err == nil {
			// Successfully loaded
			yamlLoaded = true
			break
		} else if !os.IsNotExist(err) {
			// If error is not just "file doesn't exist", return it
			return false, fmt.Errorf("error loading YAML config from %s: %w", path, err)
		}
	}

	return yamlLoaded, nil
}

// loadYAMLConfig loads configuration from a YAML file
func loadYAMLConfig(config *Config, filePath string) error {
	// Check if file exists
	_, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	// Parse YAML
	err = yaml.Unmarshal(data, config)
	if err != nil {
		return err
	}

	return nil
}

// SaveConfigToYAML saves the current configuration to a YAML file
func SaveConfigToYAML(config *Config, filePath string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if dir != "." {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return fmt.Errorf("failed to create directory for config file: %w", err)
		}
	}

	// Marshal to YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config to YAML: %w", err)
	}

	// Write to file
	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write config to file: %w", err)
	}

	return nil
}
