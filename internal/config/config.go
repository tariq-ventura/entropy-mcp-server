package config

import (
	"errors"
	"os"
	"strings"
)

type Config struct {
	Port                  string
	LogisticsServiceURL   string
	LogisticsServiceToken string
	MCPAPIKey             string
}

func Load() (Config, error) {
	config := Config{
		Port:                  envOrDefault("PORT", "3002"),
		LogisticsServiceURL:   strings.TrimRight(os.Getenv("LOGISTIC_SERVICE_URL"), "/"),
		LogisticsServiceToken: strings.TrimSpace(os.Getenv("LOGISTIC_SERVICE_TOKEN")),
		MCPAPIKey:             strings.TrimSpace(os.Getenv("MCP_API_KEY")),
	}

	if config.LogisticsServiceURL == "" {
		return Config{}, errors.New(
			"LOGISTIC_SERVICE_URL is required",
		)
	}

	if config.MCPAPIKey == "" {
		return Config{}, errors.New(
			"MCP_API_KEY is required",
		)
	}

	return config, nil
}

func envOrDefault(name string, defaultValue string) string {
	value := strings.TrimSpace(os.Getenv(name))

	if value == "" {
		return defaultValue
	}

	return value
}
