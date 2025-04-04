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

// AddBookmarksToNotion adds multiple bookmarks to Notion in a batch using the global client
func AddBookmarksToNotion(databaseID string, bookmarks []kobo.Bookmark) error {
	if defaultService == nil {
		return errors.New(ErrNotionClientNotInitialized)
	}
	return defaultService.AddBookmarks(databaseID, bookmarks)
}

func (s *NotionService) ArchivePage(databaseID string, pageID notionapi.PageID) (error) {
	_, err := s.pageClient.Update(s.contextFunc(), pageID, &notionapi.PageUpdateRequest{
		Archived: true, 
		Properties: nil, 
		Icon: nil, 
		Cover: nil,
	})

	if err != nil {
		return err
	}

	return nil
}