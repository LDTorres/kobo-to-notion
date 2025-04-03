package main

import (
	"kobo-to-notion/config"
	"kobo-to-notion/kobo"
	"kobo-to-notion/logger"
	"kobo-to-notion/notion"
	"kobo-to-notion/utils"
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
	err = notion.InitializeNotionClient(appConfig.CertPath, appConfig.NotionToken, appConfig.DatabaseID, appConfig.CreateIndividualBookmarks)
	if err != nil {
		logger.Logger.Fatalf("Error initializing Notion client: %v", err)
	}

	// Process bookmarks based on configuration
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
	logger.Logger.Printf("Creating bookmarks individually: %v\n", config.CreateIndividualBookmarks)
	
	// Fetch bookmarks from Kobo database
	bookmarks, err := kobo.GetBookmarks(config.DBPath)
	if err != nil {
		logger.Logger.Fatalf("Error retrieving highlights from database: %v", err)
	}
	
	if config.CreateIndividualBookmarks {
		processIndividualBookmarks(config.DatabaseID, bookmarks)
	} else {
		processGroupedBookmarks(config.DatabaseID, bookmarks)
	}
}

// Process bookmarks individually 
func processIndividualBookmarks(databaseID string, bookmarks []kobo.Bookmark) {
	// Get existing bookmark IDs
	existingBookmarks, err := notion.GetNotionBookmarkIDs(databaseID)
	if err != nil {
		logger.Logger.Fatalf("Error fetching Notion bookmarks: %v", err)
	}
	logger.Logger.Println("Existing bookmarks:", len(existingBookmarks))
	
	// Filter new highlights
	newBookmarks := utils.FilterNewBookmarks(bookmarks, existingBookmarks)
	logger.Logger.Printf("New bookmarks to add individually: %d\n", len(newBookmarks))
	
	// Add to Notion
	err = notion.AddBookmarksToNotion(databaseID, newBookmarks)
	if err != nil {
		logger.Logger.Fatalf("Error adding bookmarks to Notion: %v", err)
	}
}

// Process bookmarks grouped by book
func processGroupedBookmarks(databaseID string, bookmarks []kobo.Bookmark) {
	// Get existing book pages
	_, err := notion.GetBookPagesByName(databaseID)
	if err != nil {
		logger.Logger.Fatalf("Error fetching Notion book pages: %v", err)
	}
	
	logger.Logger.Printf("Processing %d bookmarks in grouped mode\n", len(bookmarks))
	
	// Add to Notion
	err = notion.AddBookmarksToNotion(databaseID, bookmarks)
	if err != nil {
		logger.Logger.Fatalf("Error adding bookmarks to Notion: %v", err)
	}
}