package config

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
)

// EnvLoader defines an interface for environment variable loading
type EnvLoader interface {
	Load() error
	GetEnv(key string) string
}

// DefaultEnvLoader implements EnvLoader with standard os and godotenv functionality
type DefaultEnvLoader struct{}

// Load environment variables from .env file
func (d *DefaultEnvLoader) Load() error {
	return godotenv.Load()
}

// GetEnv returns an environment variable value
func (d *DefaultEnvLoader) GetEnv(key string) string {
	return os.Getenv(key)
}

// Config holds all configuration values
type Config struct {
	NotionToken string
	DatabaseID  string
	DBPath      string
	CertPath    string
}

// LoadEnv loads environment variables from .env file
func LoadEnv() error {
	return godotenv.Load()
}

// GetConfig retrieves configuration values from environment variables
func GetConfig() (Config, error) {
	return GetConfigWithLoader(&DefaultEnvLoader{})
}

// GetConfigWithLoader retrieves configuration using the provided EnvLoader
// This allows for dependency injection during testing
func GetConfigWithLoader(loader EnvLoader) (Config, error) {
	notionToken := loader.GetEnv("NOTION_TOKEN")
	databaseID := loader.GetEnv("NOTION_DATABASE_ID")
	dbPath := loader.GetEnv("KOBO_DB_PATH")
	certPath := loader.GetEnv("CERT_PATH")

	if notionToken == "" || databaseID == "" || dbPath == "" {
		return Config{}, errors.New("missing required environment variables")
	}

	return Config{
		NotionToken: notionToken,
		DatabaseID:  databaseID,
		DBPath:      dbPath,
		CertPath:    certPath,
	}, nil
}
