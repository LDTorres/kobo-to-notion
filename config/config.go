package config

import (
	"kobo-to-notion/logger"
	"os"

	"github.com/joho/godotenv"
)

// Load environment variables
func LoadEnv() error {
	return godotenv.Load("env")
}

// Get configuration values
func GetConfig() (string, string, string, string) {
	notionToken := os.Getenv("NOTION_TOKEN")
	databaseID := os.Getenv("NOTION_DATABASE_ID")
	dbPath := os.Getenv("KOBO_DB_PATH")
	certPath := os.Getenv("CERT_PATH")

	if notionToken == "" || databaseID == "" || dbPath == "" || certPath == "" {
		logger.Logger.Fatal("Missing environment variables")
	}

	return notionToken, databaseID, dbPath, certPath
}