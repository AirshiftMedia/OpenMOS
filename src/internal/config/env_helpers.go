package config

import (
	"os"
	"strconv"
	"time"
)

// Helper functions to get environment variables with defaults
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}

	value, err := time.ParseDuration(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

func getEnvAsFloat(key string, defaultValue float64) float64 {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return defaultValue
	}

	return value
}

// Default value helpers
func getDefaultString(current, defaultValue string) string {
	if current == "" {
		return defaultValue
	}
	return current
}

func getDefaultInt(current, defaultValue int) int {
	if current == 0 {
		return defaultValue
	}
	return current
}

func getDefaultBool(current, defaultValue bool) bool {
	// Zero value for bool is false, but can't determine if it's default without context
	// We'll just use the provided default
	return defaultValue
}

func getDefaultFloat(current, defaultValue float64) float64 {
	if current == 0 {
		return defaultValue
	}
	return current
}

func getDefaultDuration(current, defaultValue time.Duration) time.Duration {
	if current == 0 {
		return defaultValue
	}
	return current
}
