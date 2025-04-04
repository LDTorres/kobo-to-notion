package main

import (
	"kobo-to-notion/config"
	"kobo-to-notion/kobo"
	"kobo-to-notion/logger"
	"kobo-to-notion/notion"
)

func main() {
	err := initLogger()
	if err != nil {
		logger.Logger.Fatalf("Failed to initialize: %v", err)
	}
	defer logger.Close()

	// Load configuration
	appConfig, err := loadConfiguration()
	if err != nil {
		logger.Logger.Fatalf("Error loading configuration: %v", err)
	}

	// Initialize Notion client
	err = notion.InitializeNotionClient(appConfig.CertPath, appConfig.NotionToken, appConfig.DatabaseID)
	if err != nil {
		logger.Logger.Fatalf("Error initializing Notion client: %v", err)
	}

	// Process bookmarks
	processBookmarks(appConfig)
}

// Initialize the logger
func initLogger() error {
	return logger.Init("./logs/app.log")
}

// Load configuration from environment
func loadConfiguration() (config.Config, error) {
	err := config.LoadEnv()
	if err != nil {
		return config.Config{}, err
	}

	return config.GetConfig()
}

// Fetch data and process bookmarks
func processBookmarks(config config.Config) {
	// Fetch bookmarks from Kobo database
	bookmarks, err := kobo.GetBookmarks(config.DBPath)
	if err != nil {
		logger.Logger.Fatalf("Error retrieving highlights from database: %v", err)
	}

	// Process bookmarks
	processGroupedBookmarks(config.DatabaseID, bookmarks)
}

// Process bookmarks grouped by book
func processGroupedBookmarks(databaseID string, bookmarks []kobo.Bookmark) {
	logger.Logger.Printf("Processing %d bookmarks in grouped mode\n", len(bookmarks))

	// Add to Notion
	err := notion.AddBookmarksToNotion(databaseID, bookmarks)
	if err != nil {
		logger.Logger.Fatalf("Error adding bookmarks to Notion: %v", err)
	}
}
