package config

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Port                  string
	LogisticsServiceURL   string
	LogisticsServiceToken string
	FleetServiceURL       string
	FleetServiceToken     string
	MCPAPIKey             string
	MCPDatabaseString     string
	DBMaxOpenConns        int
	DBMaxIdleConns        int
	DBConnMaxLifetime     time.Duration
	IntegrationAPIKey     string
	UIAllowedOrigin       string
	SyncInterval          time.Duration
	EmbeddingProvider     string
	EmbeddingDimensions   int
	OllamaURL             string
	OllamaEmbeddingModel  string
	GCPProjectID          string
	VertexLocation        string
	EmbeddingModel        string
}

func Load() (Config, error) {
	dimensions, err := strconv.Atoi(envOrDefault("EMBEDDING_DIMENSIONS", "768"))
	if err != nil || dimensions <= 0 {
		return Config{}, errors.New("EMBEDDING_DIMENSIONS must be a positive integer")
	}
	syncInterval, err := time.ParseDuration(envOrDefault("SYNC_INTERVAL", "2m"))
	if err != nil || syncInterval <= 0 {
		return Config{}, errors.New("SYNC_INTERVAL must be a positive Go duration, for example 2m")
	}
	maxOpenConns, err := positiveInteger("DB_MAX_OPEN_CONNS", 10)
	if err != nil {
		return Config{}, err
	}
	maxIdleConns, err := nonNegativeInteger("DB_MAX_IDLE_CONNS", 5)
	if err != nil {
		return Config{}, err
	}
	dbConnMaxLifetime, err := time.ParseDuration(envOrDefault("DB_CONN_MAX_LIFETIME", "30m"))
	if err != nil || dbConnMaxLifetime <= 0 {
		return Config{}, errors.New("DB_CONN_MAX_LIFETIME must be a positive Go duration, for example 30m")
	}
	cfg := Config{
		Port:                  envOrDefault("PORT", "3002"),
		LogisticsServiceURL:   strings.TrimRight(os.Getenv("LOGISTIC_SERVICE_URL"), "/"),
		LogisticsServiceToken: strings.TrimSpace(os.Getenv("LOGISTIC_SERVICE_TOKEN")),
		FleetServiceURL:       strings.TrimRight(os.Getenv("FLEET_SERVICE_URL"), "/"),
		FleetServiceToken:     strings.TrimSpace(os.Getenv("FLEET_SERVICE_TOKEN")),
		MCPAPIKey:             strings.TrimSpace(os.Getenv("MCP_API_KEY")),
		MCPDatabaseString:     strings.TrimSpace(os.Getenv("MCP_DB_STRING")),
		DBMaxOpenConns:        maxOpenConns,
		DBMaxIdleConns:        maxIdleConns,
		DBConnMaxLifetime:     dbConnMaxLifetime,
		IntegrationAPIKey:     strings.TrimSpace(os.Getenv("INTEGRATION_API_KEY")),
		UIAllowedOrigin:       strings.TrimSpace(os.Getenv("UI_ALLOWED_ORIGIN")),
		SyncInterval:          syncInterval,
		EmbeddingProvider:     strings.ToLower(envOrDefault("EMBEDDING_PROVIDER", "ollama")),
		EmbeddingDimensions:   dimensions,
		OllamaURL:             strings.TrimRight(os.Getenv("OLLAMA_URL"), "/"),
		OllamaEmbeddingModel:  strings.TrimSpace(os.Getenv("OLLAMA_EMBEDDING_MODEL")),
		GCPProjectID:          strings.TrimSpace(os.Getenv("GCP_PROJECT_ID")),
		VertexLocation:        strings.TrimSpace(os.Getenv("VERTEX_LOCATION")),
		EmbeddingModel:        strings.TrimSpace(os.Getenv("EMBEDDING_MODEL")),
	}
	if cfg.LogisticsServiceURL == "" {
		return Config{}, errors.New("LOGISTIC_SERVICE_URL is required")
	}
	if cfg.FleetServiceURL == "" {
		return Config{}, errors.New("FLEET_SERVICE_URL is required")
	}
	if cfg.MCPAPIKey == "" {
		return Config{}, errors.New("MCP_API_KEY is required")
	}
	if cfg.MCPDatabaseString == "" {
		return Config{}, errors.New("MCP_DB_STRING is required")
	}
	if cfg.DBMaxIdleConns > cfg.DBMaxOpenConns {
		return Config{}, errors.New("DB_MAX_IDLE_CONNS cannot be greater than DB_MAX_OPEN_CONNS")
	}
	if cfg.IntegrationAPIKey == "" {
		cfg.IntegrationAPIKey = cfg.MCPAPIKey
	}
	if cfg.EmbeddingProvider != "vertex" && cfg.EmbeddingProvider != "ollama" {
		return Config{}, errors.New("EMBEDDING_PROVIDER must be vertex or ollama")
	}
	return cfg, nil
}

func positiveInteger(name string, defaultValue int) (int, error) {
	value, err := strconv.Atoi(envOrDefault(name, strconv.Itoa(defaultValue)))
	if err != nil || value <= 0 {
		return 0, errors.New(name + " must be a positive integer")
	}
	return value, nil
}

func nonNegativeInteger(name string, defaultValue int) (int, error) {
	value, err := strconv.Atoi(envOrDefault(name, strconv.Itoa(defaultValue)))
	if err != nil || value < 0 {
		return 0, errors.New(name + " must be zero or a positive integer")
	}
	return value, nil
}
func envOrDefault(name, defaultValue string) string {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return defaultValue
	}
	return value
}
