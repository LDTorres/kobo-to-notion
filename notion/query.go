package notion

import (
	"github.com/jomei/notionapi"
)

// GetBookmarkIDs fetches all BookmarkIDs from Notion
func (s *NotionService) GetBookmarkIDs(databaseID string) (map[string]bool, error) {
	existingBookmarks := make(map[string]bool)
	var startCursor notionapi.Cursor

	for {
		query := &notionapi.DatabaseQueryRequest{}
		if startCursor != "" {
			query.StartCursor = startCursor
		}

		res, err := s.dbClient.Query(s.contextFunc(), notionapi.DatabaseID(databaseID), query)
		if err != nil {
			return nil, err
		}

		s.extractBookmarkIDs(res, existingBookmarks)

		if !res.HasMore || res.NextCursor == "" {
			break
		}
		startCursor = res.NextCursor
	}

	return existingBookmarks, nil
}

// extractBookmarkIDs extracts bookmark IDs from query results and adds them to the map
func (s *NotionService) extractBookmarkIDs(res *notionapi.DatabaseQueryResponse, bookmarks map[string]bool) {
	for _, page := range res.Results {
		if prop, ok := page.Properties[PropBookmarkID].(*notionapi.RichTextProperty); ok {
			for _, text := range prop.RichText {
				bookmarks[text.PlainText] = true
			}
		}
	}
}

// GetPagesByBookName fetches pages grouped by book name
func (s *NotionService) GetPagesByBookName(databaseID string) (map[string]notionapi.PageID, error) {
	bookPages := make(map[string]notionapi.PageID)
	var startCursor notionapi.Cursor

	for {
		query := &notionapi.DatabaseQueryRequest{}
		if startCursor != "" {
			query.StartCursor = startCursor
		}

		res, err := s.dbClient.Query(s.contextFunc(), notionapi.DatabaseID(databaseID), query)
		if err != nil {
			return nil, err
		}

		for _, page := range res.Results {
			// Extract book name from Book Title property
			if titleProp, ok := page.Properties[PropBookTitle].(*notionapi.TitleProperty); ok && len(titleProp.Title) > 0 {
				bookName := titleProp.Title[0].PlainText
				bookPages[bookName] = notionapi.PageID(page.ID)
			}
		}

		if !res.HasMore || res.NextCursor == "" {
			break
		}
		startCursor = res.NextCursor
	}

	return bookPages, nil
}
