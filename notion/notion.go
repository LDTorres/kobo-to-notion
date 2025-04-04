package notion

/*
Package notion provides functionality to interact with the Notion API.
This package is divided into several files:
- core.go: Constants, interfaces and basic functions
- query.go: Notion database queries
- blocks.go: Content block manipulation
- add_grouped.go: Add bookmarks grouped by books
*/

// This file serves as an entry point and re-exports the package's functionality
import (
	"errors"
	"kobo-to-notion/kobo"

	"github.com/jomei/notionapi"
)

// GetNotionBookmarkIDs fetches all BookmarkIDs from Notion using the global client
func GetNotionBookmarkIDs(databaseID string) (map[string]bool, error) {
	if defaultService == nil {
		return nil, errors.New(ErrNotionClientNotInitialized)
	}
	return defaultService.GetBookmarkIDs(databaseID)
}

// GetBookPagesByName fetches all book pages from Notion using the global client
func GetBookPagesByName(databaseID string) (map[string]notionapi.PageID, error) {
	if defaultService == nil {
		return nil, errors.New(ErrNotionClientNotInitialized)
	}
	return defaultService.GetPagesByBookName(databaseID)
}

// AddBookmarkToNotion adds a new bookmark to Notion using the global client
func AddBookmarkToNotion(databaseID string, bookmark kobo.Bookmark) (*notionapi.Page, error) {
	if defaultService == nil {
		return nil, errors.New(ErrNotionClientNotInitialized)
	}
	return defaultService.AddBookmark(databaseID, bookmark)
}

// AddBookmarksToNotion adds multiple bookmarks to Notion in a batch using the global client
func AddBookmarksToNotion(databaseID string, bookmarks []kobo.Bookmark) error {
	if defaultService == nil {
		return errors.New(ErrNotionClientNotInitialized)
	}
	return defaultService.AddBookmarks(databaseID, bookmarks)
}
