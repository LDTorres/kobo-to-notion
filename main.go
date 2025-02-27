package main

import (
	"kobo-to-notion/config"
	"kobo-to-notion/kobo"
	"kobo-to-notion/logger"
	"kobo-to-notion/notion"
	"kobo-to-notion/utils"

	"github.com/jomei/notionapi"
)

func main() {
	err := logger.Init("./logs/app.log")
	if err != nil {
		logger.Logger.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Close()

	// Load environment variables
	err = config.LoadEnv()
	if err != nil {
		logger.Logger.Fatalf("Error loading .env file: %v", err)
	}

	// Retrieve credentials
	notionToken, databaseID, dbPath, certPath := config.GetConfig()

	// Configure a secure HTTP client with embedded certificates
	httpClient, err := utils.ConfigureSecureHTTPClientWithFile(certPath)
	if err != nil {
		logger.Logger.Fatalf("Error configuring secure HTTP client: %v", err)
	}
	
	// Initialize Notion client
	client := notionapi.NewClient(notionapi.Token(notionToken), notionapi.WithHTTPClient(httpClient))

	// Fetch existing bookmarks from Notion
	existingBookmarks, err := notion.GetNotionBookmarkIDs(client, databaseID)

	if err != nil {
		logger.Logger.Fatalf("Error fetching Notion bookmarks: %v", err)
	}
	
	logger.Logger.Println("Existing bookmarks:", len(existingBookmarks))

	// Fetch new highlights from Kobo database
	bookmarks, err := kobo.GetBookmarks(dbPath)

	if err != nil {
		logger.Logger.Fatalf("Error retrieving highlights from database: %v", err)
	}

	// Filter new highlights (only those not in Notion)
	newBookmarks := utils.FilterNewBookmarks(bookmarks, existingBookmarks)

	logger.Logger.Printf("New bookmarks to add: %d\n", len(newBookmarks))

	// Add new highlights to Notion
	for _, bookmark := range newBookmarks {
		logger.Logger.Printf("Adding Bookmark: %s\n", bookmark.BookmarkID)

		err := notion.AddBookmarkToNotion(client, databaseID, bookmark)
		if err != nil {
			logger.Logger.Printf("Error adding highlight to Notion: %v", err)
		} else {
			logger.Logger.Println("Bookmark successfully added to Notion!")
		}
	}
}